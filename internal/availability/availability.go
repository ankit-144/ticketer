package availability

import "ticketer/internal/catalog"

type AvailabilityService struct{
	
}

func New() *AvailabilityService{
	return &AvailabilityService{}
} 


type Service interface {
    GetAvailableSeats(showID string) ([]catalog.Seat, error)
    LockSeats(showID string, seatIDs []string) error
    ReleaseSeats(showID string, seatIDs []string) error
}
