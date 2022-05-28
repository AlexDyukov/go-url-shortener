package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type InDatabase struct {
	db *sql.DB
}

func NewInDatabase(dsn string) (Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return NewInMemory(), err
	}

	return &InDatabase{db}, nil
}

func (idb *InDatabase) Get(ctx context.Context, sid ShortID) (FullURL, bool) {
	return FullURL(""), true
}

func (idb *InDatabase) Save(ctx context.Context, sid ShortID, furl FullURL) error {
	return nil
}

func (idb *InDatabase) Put(ctx context.Context, furl FullURL) (ShortID, error) {
	return ShortID(0), nil
}

func (idb *InDatabase) GetURLs(ctx context.Context) URLs {
	return URLs{}
}

func (idb *InDatabase) NewUser(ctx context.Context) User {
	return DefaultUser
}

func (idb *InDatabase) Ping(_ context.Context) bool {
	err := idb.db.Ping()
	return err == nil
}
