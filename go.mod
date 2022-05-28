module github.com/alexdyukov/go-url-shortener

go 1.17

require (
	github.com/caarlos0/env/v6 v6.9.1
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gorilla/mux v1.8.0
	github.com/jackc/pgx/v4 v4.16.1
	github.com/shomali11/util v0.0.0-20200329021417-91c54758c87b
	github.com/stretchr/testify v1.7.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/alexdyukov/go-url-shortener/internal/service => ./internal/service

replace github.com/alexdyukov/go-url-shortener/internal/storage => ./internal/storage

replace github.com/alexdyukov/go-url-shortener/internal/webhandler => ./internal/webhandler
