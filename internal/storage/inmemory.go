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
	ims.shorts[DefaultUserID] = URLs{}
	return &ims
}

func (ims *InMemory) Get(surl ShortURL) (FullURL, bool) {
	ims.mutex.RLock()
	userShorts := ims.shorts[DefaultUserID]
	savedURL, exist := userShorts.Get(surl)
	ims.mutex.RUnlock()

	return savedURL, exist
}

func (ims *InMemory) Save(surl ShortURL, furl FullURL) error {
	ims.mutex.Lock()
	userShorts := ims.shorts[DefaultUserID]
	err := userShorts.Save(surl, furl)
	ims.mutex.Unlock()

	return err
}

func (ims *InMemory) Put(furl FullURL) (ShortURL, error) {
	surl := short(furl)

	err := ims.Save(surl, furl)

	return surl, err
}

func (ims *InMemory) SetAuthor(surl ShortURL, furl FullURL, user User) error {
	ims.mutex.Lock()
	userShorts, exist := ims.shorts[user]
	if !exist {
		ims.shorts[user] = URLs{}
	}
	err := userShorts.Save(surl, furl)

	ims.mutex.Unlock()

	return err
}

func (ims *InMemory) GetAuthorURLs(user User) (URLs, bool) {
	ims.mutex.RLock()
	urls, exist := ims.shorts[user]
	ims.mutex.RUnlock()

	return urls, exist
}
