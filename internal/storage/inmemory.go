package storage

import (
	"sync"

	"github.com/shomali11/util/xhashes"
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
	id := ID(xhashes.FNV64a(str))

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	link, exist := ims.cache[id]
	if !exist {
		ims.cache[id] = str
		return id, nil
	}

	if link != str {
		return 0, ErrConflict{}
	}

	return id, nil
}
