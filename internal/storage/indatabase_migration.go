package storage

var pgInit = []string{
	//base
	"CREATE TABLE IF NOT EXISTS migration()",
	"CREATE SEQUENCE IF NOT EXISTS seq_user START 1;",
	//tables
	"CREATE TABLE IF NOT EXISTS urls ();",
	"CREATE TABLE IF NOT EXISTS relations ();",
	//colums
	"ALTER TABLE urls ADD COLUMN IF NOT EXISTS short_id bigint UNIQUE NOT NULL;",
	"ALTER TABLE urls ADD COLUMN IF NOT EXISTS full_url VARCHAR NOT NULL;",
	"ALTER TABLE relations ADD COLUMN IF NOT EXISTS user_id bigint NOT NULL;",
	"ALTER TABLE relations ADD COLUMN IF NOT EXISTS short_id bigint NOT NULL;",
	//indexes
	"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_urls__short_id ON urls (short_id);",
	"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_relations__user_id ON relations (user_id, short_id);",
}
