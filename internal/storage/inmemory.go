package storage

import (
	"context"
	"sync"
	"sync/atomic"
)

type InMemory struct {
	mutex      sync.RWMutex
	shorts     map[User]URLs
	usersCount uint64
}

func NewInMemory() Storage {
	ims := InMemory{mutex: sync.RWMutex{}, shorts: map[User]URLs{}, usersCount: uint64(0)}
	ims.shorts[DefaultUser] = URLs{}
	return &ims
}

func (ims *InMemory) updateUsersCount(user User) {
	oldUsersCount := atomic.LoadUint64(&ims.usersCount)
	newUsersCount := uint64(user)
	for oldUsersCount < newUsersCount {
		if atomic.CompareAndSwapUint64(&ims.usersCount, oldUsersCount, newUsersCount) {
			return
		}
		oldUsersCount = atomic.LoadUint64(&ims.usersCount)
	}
}

func (ims *InMemory) Get(_ context.Context, sid ShortID) (FullURL, bool) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	userShorts := ims.shorts[DefaultUser]
	return userShorts.Get(sid)
}

func (ims *InMemory) Save(ctx context.Context, sid ShortID, furl FullURL) error {
	user, err := GetUser(ctx)
	if err != nil {
		return err
	}

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	userShorts, exists := ims.shorts[user]
	if !exists {
		userShorts = URLs{}
		ims.shorts[user] = userShorts
		go ims.updateUsersCount(user)
	}

	if err := userShorts.Save(sid, furl); err != nil {
		return err
	}

	defaultShorts := ims.shorts[DefaultUser]
	if err := defaultShorts.Save(sid, furl); err != nil {
		return err
	}

	return nil
}

func (ims *InMemory) Put(ctx context.Context, furl FullURL) (ShortID, error) {
	sid := short(furl)

	return sid, ims.Save(ctx, sid, furl)
}

func (ims *InMemory) GetURLs(ctx context.Context) URLs {
	user, err := GetUser(ctx)
	if err != nil || user == DefaultUser {
		return URLs{}
	}

	ims.mutex.RLock()
	result := URLs{}
	for sid, furl := range ims.shorts[user] {
		result[sid] = furl
	}
	ims.mutex.RUnlock()

	return result
}

func (ims *InMemory) NewUser(ctx context.Context) User {
	oldUsersCount := atomic.LoadUint64(&ims.usersCount)
	newUsersCount := oldUsersCount + 1
	for !atomic.CompareAndSwapUint64(&ims.usersCount, oldUsersCount, newUsersCount) {
		oldUsersCount = atomic.LoadUint64(&ims.usersCount)
		newUsersCount = oldUsersCount + 1
	}

	return User(newUsersCount)
}

func (ims *InMemory) Ping(_ context.Context) bool {
	return true
}
