package users

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type (
	workerDetails struct {
		Email string
	}

	Workers struct {
		guard   sync.RWMutex
		storage map[string]workerDetails
	}
)

func NewWorkers() *Workers {
	return &Workers{
		storage: make(map[string]workerDetails),
	}
}

func (u *Workers) Add(userId string, email string) error {
	u.guard.Lock()
	defer u.guard.Unlock()

	if _, ok := u.storage[userId]; ok {
		return fmt.Errorf(`worker user-id=%s is already known`, userId)
	}

	u.storage[userId] = workerDetails{Email: email}
	return nil
}

func (u *Workers) Remove(userId string) error {
	u.guard.Lock()
	defer u.guard.Unlock()

	if _, ok := u.storage[userId]; !ok {
		return fmt.Errorf(`worker user-id=%s is not existent`, userId)
	}
	delete(u.storage, userId)
	return nil
}

func (u *Workers) AllIds() []string {
	u.guard.RLock()
	defer u.guard.RUnlock()

	ids := make([]string, 0, len(u.storage))
	for id, _ := range u.storage {
		ids = append(ids, id)
	}

	return ids
}

func (u *Workers) RandomIds() RandomWorkerIds {
	rand.Seed(time.Now().UnixNano())
	return RandomWorkerIds{ids: u.AllIds()}
}

type RandomWorkerIds struct {
	ids []string
}

func (r RandomWorkerIds) Len() int {
	return len(r.ids)
}

func (r RandomWorkerIds) Get() string {
	return r.ids[rand.Intn(len(r.ids))]
}
