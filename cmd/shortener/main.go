package main

import (
	"log"
	"net/http"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	webhandler "github.com/alexdyukov/go-url-shortener/internal/webhandler"
)

func main() {
	conf := Config{}
	conf.Parse()

	stor := storage.NewInMemory()
	svc := service.NewURLShortener(stor)
	wh := webhandler.NewWebHandler(svc, conf.BaseURL)

	log.Fatal(http.ListenAndServe(conf.ServerAddress, wh.HTTPRouter()))
}
