package memory

import (
	"errors"
	"sync"
	"ticketer/internal/catalog"
)

type ShowSeatRepository struct {
	mu     sync.RWMutex
	showSeats  map[string]*catalog.ShowSeat
}

func NewShowSeatRepository() *ShowSeatRepository {
	return &ShowSeatRepository{
		showSeats: make(map[string]*catalog.ShowSeat),
	}
}

func (r *ShowSeatRepository) GetByID(id string) (*catalog.ShowSeat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seat, ok := r.showSeats[id]
	if !ok {
		return nil, errors.New("seat not found")
	}
	return seat, nil
}

func (r *ShowSeatRepository) GetAvailableSeats(showID string) ([]catalog.ShowSeat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []catalog.ShowSeat
	for _, s := range r.showSeats {
		if s.ShowID == showID && s.Status == catalog.ShowSeatStatusAvailable {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (r *ShowSeatRepository) Save(showSeat *catalog.ShowSeat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if showSeat.ID == "" {
		return errors.New("seat ID is required")
	}
	r.showSeats[showSeat.ID] = showSeat
	return nil
}

func (r *ShowSeatRepository) UpdateStatus(id string, status catalog.ShowSeatStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	seat, ok := r.showSeats[id]
	if !ok {
		return errors.New("seat not found")
	}
	seat.Status = status
	return nil
}

func (r *ShowSeatRepository) UpdateStatuses(ids []string, status catalog.ShowSeatStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range ids {
		seat, ok := r.showSeats[id]
		if !ok {
			return errors.New("seat not found")
		}
		seat.Status = status
	}
	return nil
}
