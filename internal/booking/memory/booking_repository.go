package memory

import (
	"errors"
	"sync"
	"ticketer/internal/booking"
)

type BookingRepository struct {
	mu       sync.RWMutex
	bookings map[string]*booking.Booking
}

func NewBookingRepository() *BookingRepository {
	return &BookingRepository{
		bookings: make(map[string]*booking.Booking),
	}
}

func (r *BookingRepository) Save(b *booking.Booking) (*booking.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if b.ID == "" {
		return nil, errors.New("booking ID is required")
	}
	if _, ok := r.bookings[b.ID]; ok {
		return nil , errors.New("booking already exists")
	}
	r.bookings[b.ID] = b
	return r.bookings[b.ID], nil
}

func (r *BookingRepository) UpdateStatus(id string, status booking.BookingStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.bookings[id]
	if !ok {
		return errors.New("booking not found")
	}
	b.Status = status
	return nil
}

func (r *BookingRepository) GetByID(id string) (*booking.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.bookings[id]
	if !ok {
		return nil, errors.New("booking not found")
	}
	return b, nil
}
