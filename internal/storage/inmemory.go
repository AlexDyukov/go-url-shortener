package storage

import (
	"context"
	"sync"
	"sync/atomic"
)

type InMemory struct {
	mutex  sync.RWMutex
	shorts map[User]SavedURLs
	users  int64
}

func NewInMemory() Storage {
	ims := InMemory{mutex: sync.RWMutex{}, shorts: map[User]SavedURLs{}, users: int64(0)}
	ims.shorts[DefaultUser] = SavedURLs{}
	return &ims
}

func (ims *InMemory) Get(_ context.Context, sid ShortID) (FullURL, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	furl, exist := ims.shorts[DefaultUser][sid]
	if !exist {
		return DefaultFullURL, ErrNotFound{}
	}
	return furl, nil
}

func (ims *InMemory) Save(ctx context.Context, sid ShortID, furl FullURL) error {
	user, err := GetUser(ctx)
	if err != nil {
		return err
	} else if user == DefaultUser {
		return ErrInvalidUser{}
	}

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	// save short to defaultUser which used for Get() method
	_, exist := ims.shorts[DefaultUser][sid]
	if exist {
		return ErrConflict{}
	}
	ims.shorts[DefaultUser][sid] = furl

	// user's shorts
	userShorts, exist := ims.shorts[user]
	if !exist {
		userShorts = SavedURLs{}
		ims.shorts[user] = userShorts
		go ims.AddUser(ctx, user)
	}
	userShorts[sid] = furl

	return nil
}

func (ims *InMemory) Put(ctx context.Context, furl FullURL) (ShortID, error) {
	sid := Short(furl)

	return sid, ims.Save(ctx, sid, furl)
}

func (ims *InMemory) PutBatch(ctx context.Context, batch BatchRequest) (BatchResponse, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return nil, err
	}

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	defaultShorts := ims.shorts[DefaultUser]
	userShorts, exist := ims.shorts[user]
	if !exist {
		userShorts = SavedURLs{}
		ims.shorts[user] = userShorts
		go ims.AddUser(ctx, user)
	}

	result := BatchResponse{}
	for corrid, furl := range batch {
		sid := Short(furl)

		savedurl, exist := defaultShorts[sid]
		if !exist {
			defaultShorts[sid] = furl
		} else if savedurl != furl {
			//just do nothing, response wont contains failed corrid
			continue
		}

		if _, exist = userShorts[sid]; !exist {
			userShorts[sid] = furl
		}

		result[corrid] = sid
	}

	return result, nil
}

func (ims *InMemory) GetURLs(ctx context.Context) (SavedURLs, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return nil, err
	} else if user == DefaultUser {
		return nil, ErrNotFound{}
	}

	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	result := SavedURLs{}
	for sid, furl := range ims.shorts[user] {
		result[sid] = furl
	}

	if len(result) == 0 {
		return nil, ErrNotFound{}
	}

	return result, nil
}

func (ims *InMemory) NewUser(_ context.Context) (User, error) {
	n := atomic.AddInt64(&ims.users, int64(1))
	return User(n), nil
}

func (ims *InMemory) AddUser(_ context.Context, user User) {
	o := atomic.LoadInt64(&ims.users)
	n := int64(user)
	// 1, 2, ... N, -N, -N+1, ... -1, 0
	for (n > o) || (n < o && n < 0) {
		if atomic.CompareAndSwapInt64(&ims.users, o, n) {
			return
		}
		o = atomic.LoadInt64(&ims.users)
	}
}

func (ims *InMemory) Ping(_ context.Context) bool {
	return true
}
