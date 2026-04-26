package lock

import (
	"fmt"
	"sync"
	"time"
)

type LockService interface {
	TryLock(key string) error
	Unlock(key string) error
}

type InMemoryLockService struct {
    mu sync.Mutex
   locks map[string]bool
}

func NewInMemoryLockService() *InMemoryLockService {
    return &InMemoryLockService{
        locks: make(map[string]bool),
    }
}

func (s *InMemoryLockService) Lock(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.locks[key] {
        return fmt.Errorf("key %s is already locked", key)
    }

    s.locks[key] = true
    return nil
}

func (s *InMemoryLockService) Unlock(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    delete(s.locks, key)
    return nil
}




