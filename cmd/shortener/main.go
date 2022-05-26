package main

import (
	"log"
	"net/http"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	webhandler "github.com/alexdyukov/go-url-shortener/internal/webhandler"
)

func main() {
	conf := GetConfig()

	var stor storage.Storage
	if conf.FileStoragePath != "" {
		s, err := storage.NewInFile(conf.FileStoragePath)
		if err != nil {
			log.Fatal("cannot open storage file:", err.Error())
		}
		stor = s
	} else {
		stor = storage.NewInMemory()
	}
	svc := service.NewURLShortener(stor, conf.BaseURL)
	wh := webhandler.NewWebHandler(svc, conf.EncryptKey)

	log.Fatal(http.ListenAndServe(conf.ServerAddress, wh.HTTPRouter()))
}
