package pricing

import "ticketer/internal/catalog"

type PricingService struct{}

func New() *PricingService {
	return &PricingService{}
}

type Service interface {
	CalculatePrice(movie catalog.Movie, show catalog.Show, seats []catalog.ShowSeat) (float64, error)
}

func (s *PricingService) CalculatePrice(movie catalog.Movie, show catalog.Show, seats []catalog.ShowSeat) (float64, error) {
	return movie.BasePrice * float64(len(seats)), nil	
}