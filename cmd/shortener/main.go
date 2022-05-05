package main

import (
	"fmt"
	"net/http"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	webhandler "github.com/alexdyukov/go-url-shortener/internal/webhandler"
)

func main() {
	conf := Config{}
	conf.ParseParams()

	stor := storage.NewInMemory()
	svc := service.NewURLShortener(stor)
	wh := webhandler.NewWebHandler(svc)

	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), wh.HTTPRouter())
}
