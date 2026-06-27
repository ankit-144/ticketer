package pricing

import (
	"ticketer/internal/catalog"
	"time"
)

type PricingService struct {
	theaterRepo catalog.TheaterRepository
}

func New(theaterRepo catalog.TheaterRepository) *PricingService {
	if theaterRepo == nil {
		panic("Constructor parameter is nil for New PricingService")
	}
	return &PricingService{
		theaterRepo: theaterRepo,
	}
}

type Service interface {
	CalculatePrice(movie catalog.Movie, show catalog.Show, seats []catalog.ShowSeat) (float64, error)
}

func (s *PricingService) CalculatePrice(movie catalog.Movie, show catalog.Show, seats []catalog.ShowSeat) (float64, error) {
	totalPrice := 0.0

	timeSurchargeMultiplier := 1.0
	if show.StartTime.Weekday() == time.Saturday || show.StartTime.Weekday() == time.Sunday || show.StartTime.Hour() >= 18 {
		timeSurchargeMultiplier = 1.2
	}

	seatIDs := make([]string, 0, len(seats))
	for _, s := range seats {
		seatIDs = append(seatIDs, s.SeatID)
	}

	theaterSeats, err := s.theaterRepo.GetSeatsByIDs(seatIDs)
	if err != nil {
		return 0, err
	}
	
	for _, actualSeat := range theaterSeats {
		seatPrice := movie.BasePrice
		switch actualSeat.Type {
		case catalog.SeatTypePremium:
			seatPrice *= 1.5
		case catalog.SeatTypeVIP:
			seatPrice *= 2.0
		}

		totalPrice += seatPrice * timeSurchargeMultiplier
	}

	return totalPrice, nil
}