package booking

import (
	"fmt"
	"sort"
	"time"
	"ticketer/internal/availability"
	"ticketer/internal/catalog"
	"ticketer/internal/pricing"
	"github.com/google/uuid"
)

type BookingService struct {
	availabilityService availability.Service
	bookingRepo         BookingRepository
	movieRepo           catalog.MovieRepository
	showRepo            catalog.ShowRepository
	showSeatRepo        catalog.ShowSeatRepository
	pricingService      pricing.Service
}

func NewBookingService(
	availabilityService availability.Service,
	bookingRepo BookingRepository,
	movieRepo catalog.MovieRepository,
	showRepo catalog.ShowRepository,
	showSeatRepo catalog.ShowSeatRepository,
	pricingService pricing.Service,
) *BookingService {

	if availabilityService == nil || bookingRepo == nil || movieRepo == nil || showRepo == nil || showSeatRepo == nil || pricingService == nil {
		panic("Constructor parameter is nil for NewBookingService")
	}

	return &BookingService{
		availabilityService: availabilityService,
		bookingRepo:         bookingRepo,
		movieRepo:           movieRepo,
		showRepo:            showRepo,
		showSeatRepo:        showSeatRepo,
		pricingService:      pricingService,
	}
}

type Service interface {
	InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error)
	ConfirmBooking(bookingID string) error
	CancelBooking(bookingID string) error
	RevertBooking(bookingID string) error
}

func (s *BookingService) InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error) {
	sort.Strings(seatIDs)

	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return nil, fmt.Errorf("show not found: %v", err)
	}

	movie, err := s.movieRepo.GetByID(show.MovieID)
	if err != nil {
		return nil, fmt.Errorf("movie not found: %v", err)
	}

	// acquire lock on seats using availability service
	err = s.availabilityService.LockSeats(seatIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to lock seats: %w", err)
	}
	
	err = s.availabilityService.UpdateStatuses(seatIDs,catalog.ShowSeatStatusLocked)
	if err != nil {
		return nil, fmt.Errorf("failed to lock seats: %w", err)
	}

	seats := []catalog.ShowSeat{}
	for _, seatID := range seatIDs {
		seat, err := s.showSeatRepo.GetByID(seatID)
		if err != nil {
			_ = s.availabilityService.ReleaseSeats(seatIDs)
			return nil, err
		}
		seats = append(seats, *seat)
	}

	price, err := s.pricingService.CalculatePrice(*movie, *show, seats)
	if err != nil {
		_ = s.availabilityService.ReleaseSeats(seatIDs)
		return nil, err
	}

	booking := &Booking{
		ID:               uuid.New().String(),
		UserID:           userID,
		ShowID:           showID,
		SeatIDs:          seatIDs,
		Price:            price,
		Status:           BookingStatusPending,
		CreatedTimestamp: time.Now(),
		UpdatedTimestamp: time.Now(),
	}

	booking, err = s.bookingRepo.Save(booking)
	if err != nil {
		_ = s.availabilityService.ReleaseSeats( seatIDs)
		return nil, fmt.Errorf("booking failed: %v", err)
	}

	return booking, nil
}

func (s *BookingService) ConfirmBooking(bookingID string) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %v", err)
	}

	if booking.Status != BookingStatusPending {
		return fmt.Errorf("booking is not in pending state")
	}

	// Update seat statuses to booked using availability service
	err = s.availabilityService.BookSeats(booking.SeatIDs)
	if err != nil {
		return err
	}

	// Update booking status to confirmed
	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusConfirmed)
	if err != nil {
		return err
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

	// Release locked seats
	err = s.availabilityService.ReleaseSeats(booking.SeatIDs)
	if err != nil {
		return err
	}

	// Update booking status to cancelled
	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusCancelled)
	if err != nil {
		return err
	}

	return nil
}

func (s *BookingService) CancelBooking(bookingID string) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %v", err)
	}
	if booking.Status != BookingStatusConfirmed {
		return fmt.Errorf("booking is not in confirmed state")
	}

	// Release seats back to availability pool
	err = s.availabilityService.ReleaseSeats(booking.SeatIDs)
	if err != nil {
		return err
	}

	// Update booking status to cancelled
	err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusCancelled)
	if err != nil {
		return err
	}

	return nil
}
