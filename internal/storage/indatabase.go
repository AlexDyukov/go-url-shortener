package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var pgInitMigrations []pgMigration

type InDatabase struct {
	db *sql.DB
}

func NewInDatabase(dsn string) (Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("storage: indatabase: cannot acquire DB connection:", err.Error())
	}
	conn.Close()

	for _, m := range pgInitMigrations {
		m := m
		go m.Run(ctx, db)
	}

	return &InDatabase{db}, nil
}

func (idb *InDatabase) Get(ctx context.Context, sid ShortID) (FullURL, error) {
	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return DefaultFullURL, err
	}
	defer conn.Close()

	cmd := "SELECT full_url FROM urls WHERE short_id = $1;"
	rows, err := conn.QueryContext(ctx, cmd, sid)
	if err != nil {
		return DefaultFullURL, err
	}
	defer rows.Close()

	if !rows.Next() {
		return DefaultFullURL, rows.Err()
	}
	if err = rows.Err(); err != nil {
		return DefaultFullURL, err
	}

	var furl FullURL
	err = rows.Scan(&furl)
	return furl, err
}

func (idb *InDatabase) Save(ctx context.Context, sid ShortID, furl FullURL) error {
	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cmd := "INSERT INTO urls(short_id, full_url) VALUES ($1, $2);"
	if _, err := tx.ExecContext(ctx, cmd, sid, furl); err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return err
		}
		if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return err
		}
		return ErrConflict{}
	}

	user, err := GetUser(ctx)
	if err != nil {
		return nil
	}

	cmd = "INSERT INTO relations(user_id, short_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;"
	if _, err := tx.ExecContext(ctx, cmd, user, sid); err != nil {
		return err
	}

	return tx.Commit()
}

func (idb *InDatabase) Put(ctx context.Context, furl FullURL) (ShortID, error) {
	sid := Short(furl)

	return sid, idb.Save(ctx, sid, furl)
}

func (idb *InDatabase) PutBatch(ctx context.Context, batch BatchRequest) (BatchResponse, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	cmd := "INSERT INTO urls(short_id, full_url) VALUES ($1, $2) ON CONFLICT DO NOTHING;"
	stmtURLs, err := tx.PrepareContext(ctx, cmd)
	if err != nil {
		return nil, err
	}
	defer stmtURLs.Close()

	result := BatchResponse{}
	for corrid, furl := range batch {
		sid := Short(furl)

		// empty return because of transaction rollback
		if _, err = stmtURLs.ExecContext(ctx, sid, furl); err != nil {
			return nil, err
		}

		result[corrid] = sid
	}

	if user == DefaultUser {
		return result, tx.Commit()
	}

	cmd = "INSERT INTO relations(user_id, short_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;"
	stmtRelations, err := tx.PrepareContext(ctx, cmd)
	if err != nil {
		return nil, err
	}
	defer stmtRelations.Close()

	for _, furl := range batch {
		sid := Short(furl)

		if _, err = stmtRelations.ExecContext(ctx, user, sid); err != nil {
			return nil, err
		}
	}

	return result, tx.Commit()
}

func (idb *InDatabase) GetURLs(ctx context.Context) (SavedURLs, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return nil, err
	} else if user == DefaultUser {
		return nil, ErrNotFound{}
	}

	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	cmd := "SELECT u.short_id, u.full_url FROM urls u JOIN relations r ON r.short_id = u.short_id WHERE r.user_id = $1;"
	rows, err := conn.QueryContext(ctx, cmd, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := SavedURLs{}
	var sid ShortID
	var furl FullURL
	for rows.Next() {
		if err := rows.Scan(&sid, &furl); err != nil {
			return result, err
		}
		result[sid] = furl
	}
	if err = rows.Err(); err != nil {
		return result, err
	}

	if len(result) == 0 {
		return nil, ErrNotFound{}
	}

	return result, nil
}

func (idb *InDatabase) NewUser(ctx context.Context) (User, error) {
	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return DefaultUser, err
	}
	defer conn.Close()

	cmd := "SELECT nextval('seq_user');"
	rows, err := conn.QueryContext(ctx, cmd)
	if err != nil {
		return DefaultUser, err
	}
	defer rows.Close()

	if !rows.Next() {
		return DefaultUser, rows.Err()
	}
	if err = rows.Err(); err != nil {
		return DefaultUser, err
	}

	var user User
	err = rows.Scan(&user)

	return user, err
}

func (idb *InDatabase) AddUser(ctx context.Context, newUser User) {
	//backward compatibility with memory/file storage
	//do not need to implement, because of external storage
}

func (idb *InDatabase) Ping(ctx context.Context) bool {
	for _, m := range pgInitMigrations {
		if !m.isDone() {
			return false
		}
	}

	if err := idb.db.PingContext(ctx); err != nil {
		log.Println("storage: indatabase: Ping: error:", err.Error())
		return err == nil
	}

	return true
}
