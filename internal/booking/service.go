package booking

import (
	"fmt"
	"sort"
	"ticketer/internal/catalog"
	"ticketer/internal/core/lock"
	"ticketer/internal/events"
	"ticketer/internal/pricing"
	"time"
	"context"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo         BookingRepository
	movieRepo           catalog.MovieRepository
	showRepo            catalog.ShowRepository
	showSeatRepo        catalog.ShowSeatRepository
	pricingService      pricing.Service
	lockService         lock.LockService
	eventPublisher      events.EventPublisher
}

func NewBookingService(
	bookingRepo BookingRepository,
	movieRepo catalog.MovieRepository,
	showRepo catalog.ShowRepository,
	showSeatRepo catalog.ShowSeatRepository,
	pricingService pricing.Service,
	lockService lock.LockService,
	eventPublisher events.EventPublisher,
) *BookingService {

	if bookingRepo == nil || movieRepo == nil || showRepo == nil || showSeatRepo == nil || pricingService == nil || lockService == nil {
		panic("Constructor parameter is nil for NewBookingService")
	}

	return &BookingService{
		bookingRepo:         bookingRepo,
		movieRepo:           movieRepo,
		showRepo:            showRepo,
		showSeatRepo:        showSeatRepo,
		pricingService:      pricingService,
		lockService:         lockService,
		eventPublisher:      eventPublisher,
	}
}

type Service interface {
	InitiateBooking(userID string, showID string, showSeatIDs []string) (*Booking, error)
	ConfirmBooking(bookingID string) error
	CancelBooking(bookingID string) error
	RevertBooking(bookingID string) error
	GetBookingsByUser(userID string) ([]*Booking, error)
	GetBookingDetails(bookingID string) (*BookingDetails, error)
	EnrichBooking(booking *Booking) (*BookingDetails, error)
}

func (s *BookingService) InitiateBooking(userID string, showID string, showSeatIDs []string) (*Booking, error) {
	sort.Strings(showSeatIDs)

	var successfullyLockedShowSeats []string
	var bookingSuccessful bool 

	defer func ()  {
		if !bookingSuccessful && len(successfullyLockedShowSeats) > 0 {
			s.ReleaseLockedShowSeats(successfullyLockedShowSeats)
		}
	}()
	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return nil, fmt.Errorf("show not found: %w", err)
	}

	movie, err := s.movieRepo.GetByID(show.MovieID)
	if err != nil {
		return nil, fmt.Errorf("movie not found: %w", err)
	}

	showSeats, err := s.showSeatRepo.GetByIDs(showSeatIDs)
	if err != nil {
		return nil, err
	}

	for _, showSeatID := range showSeatIDs {
		err := s.lockService.TryLock(showSeatID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to lock showSeat %s: %w", showSeatID, err)
		}
		successfullyLockedShowSeats = append(successfullyLockedShowSeats, showSeatID)
	}

	if len(showSeats) != len(showSeatIDs) {
		return nil, fmt.Errorf("one or more seats not found")
	}

	for _, seat := range showSeats {
		if seat.Status != catalog.ShowSeatStatusAvailable {
			return nil, fmt.Errorf("seat %s is not available", seat.ID)
		}
	}

	err = s.showSeatRepo.UpdateStatuses(showSeatIDs, catalog.ShowSeatStatusLocked)
	if err != nil {
		return nil, fmt.Errorf("failed to update showSeat statuses: %w", err)
	}

	price, err := s.pricingService.CalculatePrice(*movie, *show, showSeats)
	if err != nil {
		return nil, err
	}

	booking := &Booking{
		ID:               uuid.New().String(),
		UserID:           userID,
		ShowID:           showID,
		SeatIDs:          showSeatIDs,
		Price:            price,
		Status:           BookingStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	booking, err = s.bookingRepo.Save(booking)
	if err != nil {
		return nil, fmt.Errorf("booking failed: %w", err)
	}
	bookingSuccessful = true
	return booking, nil
}

func (s *BookingService) ConfirmBooking(bookingID string) error {

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status != BookingStatusPending {
		return fmt.Errorf("booking is not in pending state")
	}

	err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs,catalog.ShowSeatStatusBooked)
	if err != nil {
		return err
	}

	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusConfirmed)
	if err != nil {
		return err
	}
	s.ReleaseLockedShowSeats(booking.SeatIDs)

	// Publish event
	if s.eventPublisher != nil {
		templateData := s.buildEventTemplateData(booking)
		
		event := events.GenericEventEnvelope{
			UserID:        booking.UserID,
			EventType:     "BOOKING_CONFIRMED",
			SourceService: "ticketer",
			Timestamp:     time.Now().Format(time.RFC3339),
			TemplateData:  templateData,
		}
		_ = s.eventPublisher.PublishEvent(context.Background(), event)
	}

	return nil
}

func (s *BookingService) RevertBooking(bookingID string) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %v", err)
	}

	if booking.Status != BookingStatusPending {
		return fmt.Errorf("booking is not in pending state cannot be reverted")
	}

	err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs,catalog.ShowSeatStatusAvailable)
	if err != nil {
		return err
	}
	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusReverted)
	if err != nil {
		return err
	}
	s.ReleaseLockedShowSeats(booking.SeatIDs)
	return nil
}

func (s *BookingService) CancelBooking(bookingID string) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if booking.Status != BookingStatusConfirmed {
		return fmt.Errorf("booking is not in confirmed state")
	}

	err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs,catalog.ShowSeatStatusAvailable)
	if err != nil {
		return err
	}

	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusCancelled)
	if err != nil {
		return err
	}

	// Publish event
	if s.eventPublisher != nil {
		templateData := s.buildEventTemplateData(booking)

		event := events.GenericEventEnvelope{
			UserID:        booking.UserID,
			EventType:     "BOOKING_CANCELLED",
			SourceService: "ticketer",
			Timestamp:     time.Now().Format(time.RFC3339),
			TemplateData:  templateData,
		}
		_ = s.eventPublisher.PublishEvent(context.Background(), event)
	}

	return nil
}

func (s *BookingService) ReleaseLockedShowSeats(showSeatIDs []string) {
	for _, showSeatID := range showSeatIDs {
		s.lockService.Unlock(showSeatID)
	}
}

func (s *BookingService) GetBookingsByUser(userID string) ([]*Booking, error) {
	return s.bookingRepo.GetByUserID(userID)
}

func (s *BookingService) GetBookingDetails(bookingID string) (*BookingDetails, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}
	return s.EnrichBooking(booking)
}

func (s *BookingService) EnrichBooking(booking *Booking) (*BookingDetails, error) {
	details := &BookingDetails{
		Booking: booking,
	}

	show, err := s.showRepo.GetByID(booking.ShowID)
	if err == nil {
		details.Show = show
		movie, err := s.movieRepo.GetByID(show.MovieID)
		if err == nil {
			details.Movie = movie
		}
	}

	showSeats, err := s.showSeatRepo.GetByIDs(booking.SeatIDs)
	if err == nil {
		details.ShowSeats = showSeats
	}

	return details, nil
}

func (s *BookingService) buildEventTemplateData(booking *Booking) map[string]interface{} {
	details, _ := s.EnrichBooking(booking)
	
	var movieTitle string
	var showTime string
	var seatNames []string

	if details.Movie != nil {
		movieTitle = details.Movie.Title
	}
	if details.Show != nil {
		showTime = details.Show.StartTime.Format(time.RFC3339)
	}
	for _, seat := range details.ShowSeats {
		seatNames = append(seatNames, seat.SeatID)
	}

	return map[string]any{
		"bookingId": booking.ID,
		"showId":    booking.ShowID,
		"price":     booking.Price,
		"movieName": movieTitle,
		"showTime":  showTime,
		"seats":     seatNames,
	}
}
