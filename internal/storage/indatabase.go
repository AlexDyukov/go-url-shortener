package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var pgInitMigrations []pgMigration

const maxValuesInAnyClause = 100

type toDelete struct {
	User   User
	Shorts []ShortID
}

type InDatabase struct {
	db   *sql.DB
	junk chan<- toDelete
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
		m.Run(ctx, db)
	}

	ch := make(chan toDelete)
	go func(db *sql.DB, ch <-chan toDelete) {
		cmd := "UPDATE urls SET isdeleted = true FROM relations AS r WHERE r.short_id = urls.short_id AND r.user_id = $1 AND urls.isdeleted = false AND urls.short_id = ANY ($2);"
		var val toDelete
		var user User
		var shorts []ShortID
		for {
			val = <-ch
			user = val.User
			shorts = val.Shorts

			for limit := len(shorts); limit > 0; limit = len(shorts) {
				if limit > maxValuesInAnyClause {
					limit = maxValuesInAnyClause
				}

				sToDelete := shorts[:limit]
				shorts = shorts[limit:]

				if _, err := db.Exec(cmd, user, sToDelete); err != nil {
					log.Println("storage: indatabase: background task: cannot execute update query:", err.Error())
					continue
				}
			}
		}
	}(db, ch)

	return &InDatabase{db: db, junk: ch}, nil
}

func (idb *InDatabase) Get(ctx context.Context, sid ShortID) (FullURL, error) {
	cmd := "SELECT full_url, isdeleted FROM urls WHERE short_id = $1;"
	rows, err := idb.db.QueryContext(ctx, cmd, sid)
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
	var isdeleted *bool
	if err = rows.Scan(&furl, &isdeleted); err != nil {
		return furl, err
	}
	if *isdeleted {
		return furl, ErrDeleted{}
	}
	return furl, nil
}

func (idb *InDatabase) Save(ctx context.Context, sid ShortID, furl FullURL) error {
	cmd := "INSERT INTO urls(short_id, full_url) VALUES ($1, $2);"
	if _, err := idb.db.ExecContext(ctx, cmd, sid, furl); err != nil {
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
	_, err = idb.db.ExecContext(ctx, cmd, user, sid)

	return err
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

	tx, err := idb.db.BeginTx(ctx, nil)
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

	for _, sid := range result {
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

	cmd := "SELECT u.short_id, u.full_url FROM urls u JOIN relations r ON r.short_id = u.short_id WHERE r.user_id = $1 AND NOT u.isdeleted;"
	rows, err := idb.db.QueryContext(ctx, cmd, user)
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

func (idb *InDatabase) DeleteURLs(ctx context.Context, sids []ShortID) error {
	if _, err := GetUser(ctx); err != nil {
		return err
	}

	go func() {
		_ = idb.AsyncDeleteURLs(ctx, sids)
	}()

	return nil
}

func (idb *InDatabase) AsyncDeleteURLs(ctx context.Context, sids []ShortID) []ShortID {
	user, err := GetUser(ctx)
	if err != nil {
		return sids
	}

	idb.junk <- toDelete{User: user, Shorts: sids}

	return sids
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
	//for _, m := range pgInitMigrations {
	//	if !m.isDone() {
	//		return false
	//	}
	//}

	if err := idb.db.PingContext(ctx); err != nil {
		log.Println("storage: indatabase: Ping: error:", err.Error())
		return err == nil
	}

	return true
}
