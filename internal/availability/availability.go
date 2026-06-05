package availability

import (
	"fmt"
	"ticketer/internal/catalog"
	"ticketer/internal/core/lock"
)

type AvailabilityService struct {
	showSeatRepo catalog.ShowSeatRepository
	lockService  lock.LockService
}

func New(showSeatRepo catalog.ShowSeatRepository, lockService lock.LockService) *AvailabilityService {
	if showSeatRepo == nil || lockService == nil {
		panic("Constructor parameter is nil for New AvailabilityService")
	}
	return &AvailabilityService{
		showSeatRepo: showSeatRepo,
		lockService:  lockService,
	}
}

type Service interface {
	GetAvailableSeats(showID string) ([]catalog.ShowSeat, error)
	LockSeats(showSeatIDs []string) error
	UpdateStatuses(showSeatIDs []string, status catalog.ShowSeatStatus) error
	ReleaseSeats(showSeatIDs []string) error
	BookSeats(showSeatIDs []string) error
}

func (s *AvailabilityService) GetAvailableSeats(showID string) ([]catalog.ShowSeat, error) {
	return s.showSeatRepo.GetAvailableSeats(showID)
}

// for showSeatIDs which are locked only can be booked, this is not inplemented in bookSeats
func (s *AvailabilityService) LockSeats(showSeatIDs []string) error {
	var successfullyLocked []string

	for _, showSeatID := range showSeatIDs {
		err := s.lockService.TryLock(showSeatID)
		if err != nil {
			s.releaseLocks(successfullyLocked)
			return err
		}
		
		seat, err := s.showSeatRepo.GetByID(showSeatID)
		if err != nil {
			s.releaseLocks(successfullyLocked)
			_ = s.lockService.Unlock(showSeatID)
			return err
		}
		if seat.Status != catalog.ShowSeatStatusAvailable {
			s.releaseLocks(successfullyLocked)
			_ = s.lockService.Unlock(showSeatID)
			return fmt.Errorf("seat %s is not available (current status: %s)", showSeatID, seat.Status)
		}

		successfullyLocked = append(successfullyLocked, showSeatID)
	}
	return nil
}
func (s *AvailabilityService) UpdateStatuses(showSeatIDs []string, status catalog.ShowSeatStatus) error {
   err := s.showSeatRepo.UpdateStatuses(showSeatIDs, status)
	if err != nil {
		s.releaseLocks(showSeatIDs)
		return err
	}
	return nil
}

func (s *AvailabilityService) ReleaseSeats(showSeatIDs []string) error {
  	err := s.showSeatRepo.UpdateStatuses(showSeatIDs, catalog.ShowSeatStatusAvailable)
	s.releaseLocks(showSeatIDs)
	return err
}

func (s *AvailabilityService) BookSeats(showSeatIDs []string) error {
	err := s.showSeatRepo.UpdateStatuses(showSeatIDs, catalog.ShowSeatStatusBooked)
	s.releaseLocks(showSeatIDs)
	return err
}

func (s *AvailabilityService) releaseLocks(showSeatIDs []string) {
	for _, showSeatID := range showSeatIDs {
		_ = s.lockService.Unlock(showSeatID)
	}
}

