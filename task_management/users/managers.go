package users

import (
	"fmt"
	"sync"
)

type (
	Managers struct {
		guard   sync.RWMutex
		storage map[string]struct{}
	}
)

func NewManagers() *Managers {
	return &Managers{
		storage: make(map[string]struct{}),
	}
}

func (m *Managers) Add(userId string) error {
	m.guard.Lock()
	defer m.guard.Unlock()

	if _, ok := m.storage[userId]; ok {
		return fmt.Errorf(`manager user-id=%s is already known`, userId)
	}

	m.storage[userId] = struct{}{}
	return nil
}

func (m *Managers) Remove(userId string) error {
	m.guard.Lock()
	defer m.guard.Unlock()

	if _, ok := m.storage[userId]; !ok {
		return fmt.Errorf(`manager user-id=%s is not existent`, userId)
	}
	delete(m.storage, userId)
	return nil
}

func (m *Managers) Has(userId string) bool {
	m.guard.Lock()
	defer m.guard.Unlock()

	_, ok := m.storage[userId]
	return ok
}
