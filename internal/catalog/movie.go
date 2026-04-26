package catalog

import "time"

type Movie struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"` // in minutes
	ReleaseDate time.Time `json:"release_date"`
	Genre       string    `json:"genre"`
	BasePrice   float64   `json:"base_price"`
}

type Show struct {
	ID        string    `json:"id"`
	MovieID   string    `json:"movie_id"`
	ScreenID  string    `json:"screen_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type ShowSeatStatus string
const (
	ShowSeatStatusAvailable ShowSeatStatus = "AVAILABLE"
	ShowSeatStatusLocked ShowSeatStatus = "LOCKED"
	ShowSeatStatusBooked ShowSeatStatus = "BOOKED"
)

type ShowSeat struct {
	ID string `json:"id"`
	ShowID string `json:"show_id"`
	SeatID string `json:"seat_id"`
	Price float64 `json:"price"`
	Status ShowSeatStatus `json:"status"`
}

type MovieRepository interface {
	GetByID(id string) (*Movie, error)
	List() ([]Movie, error)
	Save(movie *Movie) error
}

type ShowRepository interface {
	GetByID(id string) (*Show, error)
	GetByMovie(movieID string) ([]Show, error)
	GetByTheater(theaterID string) ([]Show, error)
	Save(show *Show) error
}

type ShowSeatRepository interface {
	GetByID(id string) (*ShowSeat, error)
	GetByShow(showID string) ([]ShowSeat, error)
	Save(showSeat *ShowSeat) error
	UpdateStatus(id string, status ShowSeatStatus) error
	UpdateStatuses(ids []string, status ShowSeatStatus) error
	GetAvailableSeats(showID string) (*ShowSeat, error)
}
