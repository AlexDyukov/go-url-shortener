package storage

import (
	"sync"
)

type InMemory struct {
	mutex  sync.RWMutex
	shorts map[User]URLs
}

func NewInMemory() Storage {
	ims := InMemory{sync.RWMutex{}, map[User]URLs{}}
	ims.shorts[DefaultUser] = URLs{}
	return &ims
}

func (ims *InMemory) Get(_ User, sid ShortID) (FullURL, bool) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	userShorts := ims.shorts[DefaultUser]

	return userShorts.Get(sid)
}

func (ims *InMemory) Save(user User, sid ShortID, furl FullURL) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	userShorts, exists := ims.shorts[user]
	if !exists {
		userShorts = URLs{}
		ims.shorts[user] = userShorts
	}

	if err := userShorts.Save(sid, furl); err != nil {
		return err
	}

	defaultShorts := ims.shorts[DefaultUser]
	return defaultShorts.Save(sid, furl)
}

func (ims *InMemory) Put(user User, furl FullURL) (ShortID, error) {
	sid := short(furl)

	err := ims.Save(user, sid, furl)

	return sid, err
}
func (ims *InMemory) GetURLs(user User) URLs {
	ims.mutex.RLock()
	result := URLs{}
	for sid, furl := range ims.shorts[user] {
		result[sid] = furl
	}
	ims.mutex.RUnlock()

	return result
}
