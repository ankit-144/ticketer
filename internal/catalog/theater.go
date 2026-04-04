package catalog

type Theater struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Location string   `json:"location"`
	Screens  []Screen `json:"screens"`
}

type Screen struct {
	ID        string `json:"id"`
	TheaterID string `json:"theater_id"`
	Name      string `json:"name"`
	Seats     []Seat `json:"seats"`
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
	GetScreen(screenID string) (*Screen, error)
	List() ([]Theater, error)
	AddSeatToScreen(screenID string, seat *Seat) (*Screen, error)
}
