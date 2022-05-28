package storage

import "context"

type UserCtxKey struct{}

type Storage interface {
	Get(ctx context.Context, sid ShortID) (FullURL, bool)
	Save(ctx context.Context, sid ShortID, furl FullURL) error
	Put(ctx context.Context, furl FullURL) (ShortID, error)
	GetURLs(ctx context.Context) URLs
	NewUser(ctx context.Context) User
	Ping(ctx context.Context) bool
}
