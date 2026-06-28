package booking

import (
	"time"
	"ticketer/internal/catalog"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
	BookingStatusReverted  BookingStatus = "REVERTED"
)

type Booking struct {
	ID        string        `json:"id"`
	ShowID    string        `json:"show_id"`
	UserID    string        `json:"user_id"`
	SeatIDs   []string      `json:"seat_ids"`
	Status    BookingStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Price     float64       `json:"price"`
}

type BookingDetails struct {
	Booking   *Booking           `json:"booking"`
	Movie     *catalog.Movie     `json:"movie"`
	Show      *catalog.Show      `json:"show"`
	ShowSeats []catalog.ShowSeat `json:"show_seats"`
}

type BookingRepository interface {
	Save(booking *Booking) (*Booking, error)
	UpdateStatus(id string, status BookingStatus) error
	GetByID(id string) (*Booking, error)
	GetByUserID(userID string) ([]*Booking, error)
}
