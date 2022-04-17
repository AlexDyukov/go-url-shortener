module github.com/alexdyukov/go-url-shortener

go 1.17

require github.com/julienschmidt/httprouter v1.3.0

require github.com/shomali11/util v0.0.0-20200329021417-91c54758c87b // indirect

replace github.com/alexdyukov/go-url-shortener/internal/service => ./internal/service
replace github.com/alexdyukov/go-url-shortener/internal/storage => ./internal/storage
replace github.com/alexdyukov/go-url-shortener/internal/webhandler => ./internal/webhandler

