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

type MovieRepository interface {
	GetByID(id string) (*Movie, error)
	List() ([]Movie, error)
	Save(movie *Movie) error
}

type ShowRepository interface {
	GetByID(id string) (*Show, error)
	GetByMovie(movieID string) ([]Show, error)
	GetByTheater(theaterID string) ([]Show, error)
}
