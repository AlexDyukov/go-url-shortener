package storage

import (
	"context"
	"database/sql"
	"log"
	"sync/atomic"
)

type pgMigration struct {
	name     string
	commands []string
	done     int32
}

func (m *pgMigration) Run(ctx context.Context, db *sql.DB) {
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("storage: indatabase: cannot acquire DB connection:", err.Error())
	}
	defer conn.Close()

	for i := 0; i < len(m.commands); i += 1 {
		cmd := m.commands[i]

		if _, err := conn.ExecContext(ctx, cmd); err != nil {
			log.Fatalf("storage: indatabase: cannot initialize database. Command '%s' failed with error:%s", cmd, err.Error())
		}

		atomic.AddInt32(&m.done, 1)
	}
}

func (m *pgMigration) isDone() bool {
	return int32(len(m.commands)) == atomic.LoadInt32(&m.done)
}

func (m *pgMigration) GetName() string {
	return m.name
}

func init() {
	pgInitMigrations = append(pgInitMigrations, pgMigration{
		name: "user sequence",
		commands: []string{
			"CREATE SEQUENCE IF NOT EXISTS seq_user START 1;",
		},
		done: 0,
	})
	pgInitMigrations = append(pgInitMigrations, pgMigration{
		name: "urls table",
		commands: []string{
			"CREATE TABLE IF NOT EXISTS urls ();",
			"ALTER TABLE urls ADD COLUMN IF NOT EXISTS short_id bigint UNIQUE NOT NULL;",
			"ALTER TABLE urls ADD COLUMN IF NOT EXISTS full_url VARCHAR NOT NULL;",
			//https://www.postgresql.org/docs/current/sql-createtable.html
			//PostgreSQL automatically creates an index for each unique constraint and primary key constraint to enforce uniqueness.
			//"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_urls__short_id ON urls (short_id);",
		},
		done: 0,
	})
	pgInitMigrations = append(pgInitMigrations, pgMigration{
		name: "relations table",
		commands: []string{
			"CREATE TABLE IF NOT EXISTS relations ();",
			"ALTER TABLE relations ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL;",
			"ALTER TABLE relations ADD COLUMN IF NOT EXISTS short_id bigint NOT NULL;",
			"CREATE INDEX IF NOT EXISTS idx_relations__user_id ON relations (user_id);",
			"CREATE INDEX IF NOT EXISTS idx_relations__short_id ON relations (short_id);",
		},
		done: 0,
	})
}
