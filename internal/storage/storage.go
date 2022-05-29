package storage

import "context"

type UserCtxKey struct{}

type ErrInternalError struct{}

func (e ErrInternalError) Error() string {
	return "Storage: error processing request"
}

type Storage interface {
	Get(ctx context.Context, sid ShortID) (FullURL, error)
	Save(ctx context.Context, sid ShortID, furl FullURL) error
	Put(ctx context.Context, furl FullURL) (ShortID, error)
	GetURLs(ctx context.Context) (URLs, error)
	NewUser(ctx context.Context) (User, error)
	Ping(ctx context.Context) bool
}
