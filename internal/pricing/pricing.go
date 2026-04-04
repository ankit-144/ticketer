package pricing

import "ticketer/internal/catalog"

type PricingService struct{}

func New() *PricingService {
	return &PricingService{}
}


type Service interface {
    CalculatePrice(movie catalog.Movie, show catalog.Show, seats []catalog.Seat) (float64, error)
}