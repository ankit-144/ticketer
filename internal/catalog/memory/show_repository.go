package memory

import (
	"errors"
	"sync"
	"ticketer/internal/catalog"
)

type ShowRepository struct {
	mu     sync.RWMutex
	shows  map[string]*catalog.Show
}

func NewShowRepository() *ShowRepository {
	return &ShowRepository{
		shows: make(map[string]*catalog.Show),
	}
}

func (r *ShowRepository) GetByID(id string) (*catalog.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	show, ok := r.shows[id]
	if !ok {
		return nil, errors.New("show not found")
	}
	return show, nil
}

func (r *ShowRepository) GetByMovie(movieID string) ([]catalog.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []catalog.Show
	for _, s := range r.shows {
		if s.MovieID == movieID {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (r *ShowRepository) GetByTheater(theaterID string) ([]catalog.Show, error) {
	// Note: In a real DB, we'd query by ScreenID which belongs to a Theater.
	// For memory, we'll just return all for now or filter if we had a screen-to-theater mapping.
	// This is a simplified version.
	return nil, errors.New("not implemented: requires screen-to-theater mapping")
}

func (r *ShowRepository) Save(show *catalog.Show) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if show.ID == "" {
		return errors.New("show ID is required")
	}
	r.shows[show.ID] = show
	return nil
}


