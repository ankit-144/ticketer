package catalog

type Theater struct {
	ID       string   `json:"id"`
	AdminID  string   `json:"admin_id"`
	Name     string   `json:"name"`
	Location string   `json:"location"`
}

type Screen struct {
	ID        string `json:"id"`
	TheaterID string `json:"theater_id"`
	Name      string `json:"name"`
}

type SeatType string

const (
	SeatTypeNormal  SeatType = "NORMAL"
	SeatTypePremium SeatType = "PREMIUM"
	SeatTypeVIP     SeatType = "VIP"
)

type Seat struct {
	ID       string   `json:"id"`
	ScreenID string   `json:"screen_id"`
	Row      string   `json:"row"`
	Number   int      `json:"number"`
	Type     SeatType `json:"type"`
}

type TheaterRepository interface {
	GetByID(id string) (*Theater, error)
	GetByAdminID(adminID string) ([]Theater, error)
	GetScreen(screenID string) (*Screen, error)
	GetScreens(theaterID string) ([]Screen, error)
	GetSeats(screenID string) ([]Seat, error)
	GetSeatsByIDs(seatIDs []string) ([]Seat, error)
	List() ([]Theater, error)
	Save(theater *Theater) error
	Update(theater *Theater) error
	Delete(id string) error
	AddScreenToTheater(theaterID string, screen *Screen) error
	UpdateScreen(screen *Screen) error
	DeleteScreen(theaterID, screenID string) error
	AddSeatToScreen(screenID string, seat *Seat) error
	UpdateSeat(seat *Seat) error
	DeleteSeat(screenID, seatID string) error
}
