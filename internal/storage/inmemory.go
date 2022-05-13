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

func (ims *InMemory) Save(id ID, str string) {
	ims.mutex.Lock()
	ims.cache[id] = str
	ims.mutex.Unlock()
}

func (ims *InMemory) Put(str string) (ID, error) {
	id := hash(str)

	link, exist := ims.Get(id)
	if exist {
		if link == str {
			return id, nil

		}
		return id, ErrConflict{}
	}

	ims.Save(id, str)
	return id, nil
}
