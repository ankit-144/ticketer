package memory

import (
	"fmt"
	"sync"
	"ticketer/internal/catalog"
)

type TheaterRepository struct{
	mu     sync.RWMutex
	theater map[string]*catalog.Theater
	screenTheaterMap map[string]string
}


func NewTheaterRepository() *TheaterRepository {
	return &TheaterRepository{
		theater:          make(map[string]*catalog.Theater),
		screenTheaterMap: make(map[string]string),
	}
}

func (tRepo *TheaterRepository) GetByID(id string) (*catalog.Theater, error) {
	tRepo.mu.RLock()
	defer tRepo.mu.RUnlock()

	if val, ok := tRepo.theater[id]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no theater found for ID: %s", id)
}

func (tRepo *TheaterRepository) GetScreen(screenID string) (*catalog.Screen, error) {
	tRepo.mu.RLock()
	defer tRepo.mu.RUnlock()

	theaterID, ok := tRepo.screenTheaterMap[screenID]
	if !ok {
		return nil, fmt.Errorf("no theater found for screen ID: %s", screenID)
	}

	theater, ok := tRepo.theater[theaterID]
	if !ok {
		return nil, fmt.Errorf("theater %s not found for screen %s", theaterID, screenID)
	}

	for _, sc := range theater.Screens {
		if sc.ID == screenID {
			return &sc, nil
		}
	}

	return nil, fmt.Errorf("no screen with ID %s found in theater %s", screenID, theaterID)
}

func (tRepo *TheaterRepository) List() ([]catalog.Theater, error) {
	tRepo.mu.RLock()
	defer tRepo.mu.RUnlock()

	theaterList := make([]catalog.Theater, 0, len(tRepo.theater))
	for _, v := range tRepo.theater {
		theaterList = append(theaterList, *v)
	}

	return theaterList, nil
}

func (tRepo *TheaterRepository) AddSeatToScreen(screenID string, seat *catalog.Seat) (*catalog.Screen, error) {
	tRepo.mu.Lock()
	defer tRepo.mu.Unlock()
	
	screen, err := tRepo.GetScreen(screenID)
	
	if err != nil {
		return  nil, fmt.Errorf("Error while getting Screen for Id: %v", screenID, err)
	}
	
	screen.Seats = append(screen.Seats, *seat)
	
	return screen, nil
}