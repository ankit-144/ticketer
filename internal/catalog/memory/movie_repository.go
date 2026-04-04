package memory

import (
	"errors"
	"sync"
	"ticketer/internal/catalog"
)

type MovieRepository struct {
	mu     sync.RWMutex
	movies map[string]*catalog.Movie
}



// NewMovieRepository initializes a new in-memory movie repository.
func NewMovieRepository() *MovieRepository {
	return &MovieRepository{
		movies: make(map[string]*catalog.Movie),
	}
}

func (r *MovieRepository) GetByID(id string) (*catalog.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	movie, ok := r.movies[id]
	if !ok {
		return nil, errors.New("movie not found")
	}
	return movie, nil
}

func (r *MovieRepository) List() ([]catalog.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	movies := make([]catalog.Movie, 0, len(r.movies))
	for _, m := range r.movies {
		movies = append(movies, *m)
	}
	return movies, nil
}

func (r *MovieRepository) Save(movie *catalog.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if movie.ID == "" {
		return errors.New("movie ID is required")
	}
	r.movies[movie.ID] = movie
	return nil
}
