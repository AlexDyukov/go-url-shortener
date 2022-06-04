package storage

import "context"

type BatchRequest map[CorrelationID]FullURL
type BatchResponse map[CorrelationID]ShortID
type SavedURLs map[ShortID]FullURL

type Storage interface {
	Get(ctx context.Context, sid ShortID) (FullURL, error)
	Save(ctx context.Context, sid ShortID, furl FullURL) error
	Put(ctx context.Context, furl FullURL) (ShortID, error)
	PutBatch(ctx context.Context, batch BatchRequest) (BatchResponse, error)
	GetURLs(ctx context.Context) (SavedURLs, error)
	AsyncDeleteURLs(ctx context.Context, sids []ShortID) []ShortID
	DeleteURLs(ctx context.Context, sids []ShortID) error
	NewUser(ctx context.Context) (User, error)
	AddUser(ctx context.Context, user User)
	Ping(ctx context.Context) bool
}
