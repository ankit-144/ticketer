package booking

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID        string        `json:"id"`
	ShowID    string        `json:"show_id"`
	UserID    string        `json:"user_id"`
	SeatIDs   []string      `json:"seat_ids"`
	Status    BookingStatus `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Price     float64       `json:"price"`
}

type BookingRepository interface {
	Save(booking *Booking) (*Booking, error)
	UpdateStatus(id string, status BookingStatus) error
	GetByID(id string) (*Booking, error)
}
