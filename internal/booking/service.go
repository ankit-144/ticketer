package booking

import (
	"fmt"
	"sort"
	"ticketer/internal/catalog"
	"ticketer/internal/core/lock"
	"ticketer/internal/pricing"
	"time"
	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo         BookingRepository
	movieRepo           catalog.MovieRepository
	showRepo            catalog.ShowRepository
	showSeatRepo        catalog.ShowSeatRepository
	pricingService      pricing.Service
	lockService         lock.LockService
}

func NewBookingService(
	bookingRepo BookingRepository,
	movieRepo catalog.MovieRepository,
	showRepo catalog.ShowRepository,
	showSeatRepo catalog.ShowSeatRepository,
	pricingService pricing.Service,
	lockService lock.LockService,
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
	}
}

type Service interface {
	InitiateBooking(userID string, showID string, showSeatIDs []string) (*Booking, error)
	ConfirmBooking(bookingID string) error
	CancelBooking(bookingID string) error
	RevertBooking(bookingID string) error
	GetBookingsByUser(userID string) ([]*Booking, error)
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

	showSeats, err := s.showSeatRepo.GetByShow(showID)

	if err != nil {
		return nil, err
	}

	showSeatsMap := make(map[string]catalog.ShowSeat)
	for _, showSeat := range showSeats {
		showSeatsMap[showSeat.ID] = showSeat
	}

	for _, showSeatID := range showSeatIDs {
		err := s.lockService.TryLock(showSeatID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to lock showSeat %s: %w", showSeatID, err)
		}
		successfullyLockedShowSeats = append(successfullyLockedShowSeats, showSeatID)
	}

	for _, showSeatID := range showSeatIDs {
		seat, ok := showSeatsMap[showSeatID]
		if !ok {
			return nil, fmt.Errorf("seat %s not found", showSeatID)
		}
		if seat.Status != catalog.ShowSeatStatusAvailable {
			return nil, fmt.Errorf("seat %s is not available", showSeatID)
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
