package storage

import (
	"context"
	"database/sql"
	"log"

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

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("storage: indatabase: cannot acquire DB connection:", err.Error())
	}
	defer conn.Close()

	// order counts
	for i := 0; i < len(pgInit); i += 1 {
		cmd := pgInit[i]
		if _, err := conn.ExecContext(ctx, cmd); err != nil {
			log.Fatalf("storage: indatabase: cannot initialize database. Command '%s' failed with error:%s", cmd, err.Error())
		}
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
		return DefaultFullURL, ErrNotFound{}
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

	cmd := "INSERT INTO urls(short_id, full_url) VALUES ($1, $2) ON CONFLICT DO NOTHING;"
	if _, err := tx.ExecContext(ctx, cmd, sid, furl); err != nil {
		return err
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
	sid := short(furl)

	return sid, idb.Save(ctx, sid, furl)
}

func (idb *InDatabase) GetURLs(ctx context.Context) (URLs, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return URLs{}, err
	}
	if user == DefaultUser {
		return URLs{}, ErrNotFound{}
	}

	conn, err := idb.db.Conn(ctx)
	if err != nil {
		return URLs{}, err
	}
	defer conn.Close()

	cmd := "SELECT u.short_id, u.full_url FROM urls u JOIN relations r ON r.short_id = u.short_id WHERE r.user_id = $1;"
	rows, err := conn.QueryContext(ctx, cmd, user)
	if err != nil {
		return URLs{}, err
	}
	defer rows.Close()

	result := URLs{}
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
		return URLs{}, ErrNotFound{}
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
		return DefaultUser, ErrInternalError{}
	}
	if err = rows.Err(); err != nil {
		return DefaultUser, err
	}
	var user User
	err = rows.Scan(&user)

	return user, err
}

func (idb *InDatabase) Ping(_ context.Context) bool {
	err := idb.db.Ping()
	if err != nil {
		log.Println("storage: indatabase: cannot ping database:", err.Error())
	}
	return err == nil
}
