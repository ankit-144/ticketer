package catalog

import (
	"errors"

	"github.com/google/uuid"
)

type TheaterService struct {
	theaterRepo TheaterRepository
}

func NewTheaterService(tr TheaterRepository) *TheaterService {
	return &TheaterService{theaterRepo: tr}
}

func (s *TheaterService) AddTheater(theater *Theater) error {
	existing, err := s.theaterRepo.GetByAdminID(theater.AdminID)
	if err == nil {
		for _, th := range existing {
			if th.Name == theater.Name && th.Location == theater.Location {
				*theater = th
				return nil
			}
		}
	}
	theater.ID = uuid.New().String()
	if theater.Screens == nil {
		theater.Screens = make([]Screen, 0)
	}
	return s.theaterRepo.Save(theater)
}

func (s *TheaterService) GetTheater(id string) (*Theater, error) {
	return s.theaterRepo.GetByID(id)
}

func (s *TheaterService) ListTheatersByAdmin(adminID string) ([]Theater, error) {
	return s.theaterRepo.GetByAdminID(adminID)
}

func (s *TheaterService) UpdateTheater(theater *Theater) error {
	existing, err := s.theaterRepo.GetByID(theater.ID)
	if err != nil {
		return err
	}

	if existing.AdminID != theater.AdminID {
		return errors.New("unauthorized: you do not have permission to modify this theater")
	}

	allTheaters, err := s.theaterRepo.List()
		for _, t := range allTheaters {
			if t.ID != theater.ID && t.Name == theater.Name && t.Location == theater.Location {
				return errors.New("conflict: a theater with this name and location already exists")
			}
		}

	theater.Screens = existing.Screens

	return s.theaterRepo.Update(theater)
}

func (s *TheaterService) DeleteTheater(id string) error {
	return s.theaterRepo.Delete(id)
}

func (s *TheaterService) AddScreenToTheater(theaterID string, screen *Screen) error {
	// Idempotency check: if the theater already has a screen with this name, return it
	th, err := s.theaterRepo.GetByID(theaterID)
	if err == nil {
		for _, existingScreen := range th.Screens {
			if existingScreen.Name == screen.Name {
				*screen = existingScreen
				return nil
			}
		}
	}

	screen.ID = uuid.New().String()
	screen.TheaterID = theaterID
	if screen.Seats == nil {
		screen.Seats = make([]Seat, 0)
	}
	return s.theaterRepo.AddScreenToTheater(theaterID, screen)
}

func (s *TheaterService) UpdateScreen(screen *Screen) error {
	return s.theaterRepo.UpdateScreen(screen)
}

func (s *TheaterService) DeleteScreen(theaterID, screenID string) error {
	return s.theaterRepo.DeleteScreen(theaterID, screenID)
}

func (s *TheaterService) GetScreens(theaterID string) ([]Screen, error) {
	th, err := s.theaterRepo.GetByID(theaterID)
	if err != nil {
		return nil, err
	}
	return th.Screens, nil
}

func (s *TheaterService) GetScreen(screenID string) (*Screen, error) {
	return s.theaterRepo.GetScreen(screenID)
}

func (s *TheaterService) AddSeatToScreen(screenID string, seat *Seat) (*Screen, error) {
	screenObj, err := s.theaterRepo.GetScreen(screenID)
	if err == nil {
		for _, existingSeat := range screenObj.Seats {
			if existingSeat.Row == seat.Row && existingSeat.Number == seat.Number {
				*seat = existingSeat
				return screenObj, nil
			}
		}
	}

	seat.ID = uuid.New().String()
	seat.ScreenID = screenID
	return s.theaterRepo.AddSeatToScreen(screenID, seat)
}

func (s *TheaterService) UpdateSeat(seat *Seat) error {
	return s.theaterRepo.UpdateSeat(seat)
}

func (s *TheaterService) DeleteSeat(screenID, seatID string) error {
	return s.theaterRepo.DeleteSeat(screenID, seatID)
}

func (s *TheaterService) GetSeats(screenID string) ([]Seat, error) {
	return s.theaterRepo.GetSeats(screenID)
}

type MovieService struct {
	movieRepo MovieRepository
}

func NewMovieService(mr MovieRepository) *MovieService {
	return &MovieService{movieRepo: mr}
}

func (s *MovieService) AddMovie(movie *Movie) error {
	existingMovies, _:= s.movieRepo.List()
		for _, m := range existingMovies {
			if m.Title == movie.Title {
				*movie = m
				return nil
			}
		}

	movie.ID = uuid.New().String()
	return s.movieRepo.Save(movie)
}

func (s *MovieService) GetMovie(id string) (*Movie, error) {
	return s.movieRepo.GetByID(id)
}

func (s *MovieService) ListMovies() ([]Movie, error) {
	return s.movieRepo.List()
}

func (s *MovieService) UpdateMovie(movie *Movie) error {
	if _, err := s.movieRepo.GetByID(movie.ID); err != nil {
		return errors.New("movie doesn't exist")
	}

	existingMovies, err := s.movieRepo.List()
	if err == nil {
		for _, m := range existingMovies {
			if m.ID != movie.ID && m.Title == movie.Title {
				return errors.New("conflict: another movie with this title already exists")
			}
		}
	}

	return s.movieRepo.Update(movie)
}

func (s *MovieService) DeleteMovie(id string) error {
	return s.movieRepo.Delete(id)
}

type ShowService struct {
	showRepo      ShowRepository
	showSeatRepo  ShowSeatRepository
	theaterRepo   TheaterRepository
}

func NewShowService(sr ShowRepository, ssr ShowSeatRepository, tr TheaterRepository) *ShowService {
	return &ShowService{
		showRepo:     sr,
		showSeatRepo: ssr,
		theaterRepo:  tr,
	}
}

func (s *ShowService) AddShow(show *Show) error {
	existingShows, err := s.showRepo.GetByScreen(show.ScreenID)
	if err == nil {
		for _, existingShow := range existingShows {
			if existingShow.MovieID == show.MovieID && existingShow.StartTime.Equal(show.StartTime) {
				*show = existingShow
				return nil
			}
		}
	}

	show.ID = uuid.New().String()
	screen, err := s.theaterRepo.GetScreen(show.ScreenID)
	if err != nil {
		return errors.New("invalid screen ID for show: " + err.Error())
	}

	err = s.showRepo.Save(show)
	if err != nil {
		return err
	}

	for _, seat := range screen.Seats {
		showSeat := &ShowSeat{
			ID:     uuid.New().String(),
			ShowID: show.ID,
			SeatID: seat.ID,
			Status: ShowSeatStatusAvailable,
		}
		if err := s.showSeatRepo.Save(showSeat); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShowService) GetShow(id string) (*Show, error) {
	return s.showRepo.GetByID(id)
}

func (s *ShowService) GetShowsByMovie(movieID string) ([]Show, error) {
	return s.showRepo.GetByMovie(movieID)
}

func (s *ShowService) GetShowsByTheater(theaterID string) ([]Show, error) {
	th, err := s.theaterRepo.GetByID(theaterID)
	if err != nil {
		return nil, err
	}

	var allShows []Show
	for _, screen := range th.Screens {
		shows, err := s.showRepo.GetByScreen(screen.ID)
		if err == nil {
			allShows = append(allShows, shows...)
		}
	}
	return allShows, nil
}

func (s *ShowService) UpdateShow(show *Show) error {
	if _, err := s.showRepo.GetByID(show.ID); err != nil {
		return errors.New("show doesn't exist")
	}

	existingShows, err := s.showRepo.GetByScreen(show.ScreenID)
	if err == nil {
		for _, existingShow := range existingShows {
			if existingShow.ID != show.ID && existingShow.StartTime.Equal(show.StartTime) {
				return errors.New("conflict: another show is already scheduled on this screen at this start time")
			}
		}
	}

	return s.showRepo.Update(show)
}

func (s *ShowService) DeleteShow(id string) error {
	return s.showRepo.Delete(id)
}
