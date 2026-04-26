package booking

import (
	"fmt"
	"sort"
	"ticketer/internal/catalog"
	"ticketer/internal/core/lock"
	"ticketer/internal/pricing"

	"github.com/google/uuid"
)

type BookingService struct {
	// Add dependencies here (e.g., repositories, pricing service)
    lockService lock.LockService
    bookingRepo BookingRepository
    movieRepo catalog.MovieRepository
    showRepo catalog.ShowRepository
    showSeatRepo catalog.ShowSeatRepository
    pricingService pricing.Service
}

func NewBookingService(lockService lock.LockService, bookingRepo BookingRepository, movieRepo catalog.MovieRepository, showRepo catalog.ShowRepository, showSeatRepo catalog.ShowSeatRepository, pricingService pricing.Service) *BookingService {
    
    if lockService == nil || bookingRepo == nil || movieRepo == nil || showRepo == nil || showSeatRepo == nil || pricingService == nil {
        panic("Constructor parameter is nil for NewBookingService")
    }

	return &BookingService{
        lockService: lockService,
        bookingRepo: bookingRepo,
        movieRepo: movieRepo,
        showRepo: showRepo,
        showSeatRepo: showSeatRepo,
        pricingService: pricingService,
    }
}

type Service interface {

	InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error)

	ConfirmBooking(bookingID string) error

	CancelBooking(bookingID string) error

	RevertBooking(bookingID string) error
}

func (s *BookingService) InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error) {
	// 1. Fetch Show & Movie (for timing, theater)
	// 2. Fetch Seats & Prices (from Pricing Service)
	// 3. Validate Seats Available (lock/check)
	// 4. Create Booking Record (status:PENDING)
	// 5. Return Token (wait for payment)

    sort.Strings(seatIDs)
    

    show , err := s.showRepo.GetByID(showID)
    if err != nil {
        return nil, fmt.Errorf("show not found: %v", err)
    }

    movie , err := s.movieRepo.GetByID(show.MovieID)
    if err != nil {
        return nil, fmt.Errorf("movie not found: %v", err)
    }

    // acquire lock on seats 
    err = s.acquireLockOnIds(seatIDs)
    if err != nil {
        return nil, err
    }

    // release lock on seats if booking fails
    defer s.releaseLockOnIds(seatIDs)

    seats := []catalog.ShowSeat{}
    for _, seatID := range seatIDs {
        seat, err := s.showSeatRepo.GetByID(seatID)
        if err != nil {
            return nil, err
        }
        if seat.Status != catalog.ShowSeatStatusAvailable {
            return nil, fmt.Errorf("seat %s is already booked", seatID)
        }
        seats = append(seats, *seat)
    }

    price, err := s.pricingService.CalculatePrice(*movie, *show, seats)
    if err != nil {
        return nil, err
    }

    // save booking at last
    booking := &Booking{
        ID: uuid.New().String(),
        UserID: userID,
        ShowID: showID,
        SeatIDs: seatIDs,
        Price: price,
        Status: BookingStatusPending,
    }

    booking , err = s.bookingRepo.Save(booking)

    if err != nil {
        return nil, fmt.Errorf("booking failed: %v", err)
    }

    // update the status of seats to booked
    err = s.showSeatRepo.UpdateStatuses(seatIDs, catalog.ShowSeatStatusLocked)
    if err != nil {
        return nil, err
    }

	return booking, nil
}

func (s *BookingService) ConfirmBooking(bookingID string) error {
	// 1. Check Payment Status
	// 2. Lock Seats (mark as BOOKED)
	// 3. Update Booking Status: CONFIRMED
	// 4. Send Ticket

    booking , err := s.bookingRepo.GetByID(bookingID)
    if err != nil {
        return fmt.Errorf("booking not found: %v", err)
    }

    if booking.Status != BookingStatusPending {
        return fmt.Errorf("booking is not in pending state")
    }

    //acquire lock on seats 
    err = s.acquireLockOnIds(booking.SeatIDs)
    if err != nil {
        return err
    }

    // release lock on seats if booking fails
    defer s.releaseLockOnIds(booking.SeatIDs)

    // update seat statuses to booked 
    err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs, catalog.ShowSeatStatusBooked)
    if err != nil {
        return err
    }

    // update booking status to confirmed 
    err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusConfirmed)
    if err != nil {
        return err
    }

	return nil
}

func (s *BookingService) RevertBooking(bookingID string) error {
    booking , err := s.bookingRepo.GetByID(bookingID)
    if err != nil {
        return fmt.Errorf("booking not found: %v", err)
    }

    // idempotency check for already cancelled bookings
    if booking.Status != BookingStatusPending {
        return fmt.Errorf("booking is not in pending state cannot be reverted")
    }

    // acquire lock on seats 
    err = s.acquireLockOnIds(booking.SeatIDs)
    if err != nil {
        return err
    }

    // release lock on seats if booking fails
    defer s.releaseLockOnIds(booking.SeatIDs)

    // update seat statuses to available
    err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs, catalog.ShowSeatStatusAvailable)
    if err != nil {
        return err
    }

    // update booking status to confirmed 
    err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusCancelled)
    if err != nil {
        return err
    }

    return nil
}

func (s *BookingService) CancelBooking(bookingID string) error {
    // booking should be confirmed in orded to be reverted 
    booking, err := s.bookingRepo.GetByID(bookingID)
    if err != nil {
        return fmt.Errorf("booking not found: %v", err)
    }
    if booking.Status != BookingStatusConfirmed {
        return fmt.Errorf("booking is not in confirmed state")
    }

    // acquire lock on seats 
    err = s.acquireLockOnIds(booking.SeatIDs)
    if err != nil {
        return err
    }

    // release lock on seats if booking fails
    defer s.releaseLockOnIds(booking.SeatIDs)

    // update seat statuses to available
    err = s.showSeatRepo.UpdateStatuses(booking.SeatIDs, catalog.ShowSeatStatusAvailable)
    if err != nil {
        return err
    }

    // update booking status to confirmed 
    err = s.bookingRepo.UpdateStatus(bookingID, BookingStatusCancelled)
    if err != nil {
        return err
    }
    
    return nil
}


func (s *BookingService) acquireLockOnIds(ids []string) error {
    var successfullyLocked []string
    // acquire lock on seats 
    for _, seatID := range ids {
        err := s.lockService.TryLock(seatID)
        if err != nil {
            s.releaseLockOnIds(successfullyLocked)
            return err
        }
        successfullyLocked = append(successfullyLocked, seatID)
    }
    return nil  
}

func (s *BookingService) releaseLockOnIds(ids []string) error {
    // release lock on seats 
    for _, seatID := range ids {
        _ = s.lockService.Unlock(seatID)
    }
    return nil  
}

