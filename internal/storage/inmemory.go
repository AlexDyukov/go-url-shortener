package storage

import (
	"sync"
)

type InMemory struct {
	mutex sync.RWMutex
	cache map[ID]string
}

func NewInMemory() Storage {
	return &InMemory{sync.RWMutex{}, map[ID]string{}}
}

func (ims *InMemory) Get(id ID) (string, bool) {
	ims.mutex.RLock()
	value, exist := ims.cache[id]
	ims.mutex.RUnlock()

	if !exist {
		return "", false
	}

	return value, true
}

func (ims *InMemory) Put(str string) (ID, error) {
	id := hash(str)

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	link, exist := ims.cache[id]
	if !exist {
		ims.cache[id] = str
		return id, nil
	}

	if link != str {
		return ID(0), ErrConflict{}
	}

	return id, nil
}
