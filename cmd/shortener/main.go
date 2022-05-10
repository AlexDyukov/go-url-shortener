package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	webhandler "github.com/alexdyukov/go-url-shortener/internal/webhandler"
)

func main() {
	conf := Config{}
	conf.Parse()

	var stor storage.Storage
	if conf.FileStoragePath != "" {
		file, err := os.OpenFile(conf.FileStoragePath, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		stor, err = storage.NewInFile(conf.FileStoragePath)
		if err != nil {
			logStr := fmt.Sprintf("cannot open storage file: %s\n", err)
			log.Fatal(logStr)
		}
	} else {
		stor = storage.NewInMemory()
	}
	svc := service.NewURLShortener(stor)
	wh := webhandler.NewWebHandler(svc, conf.BaseURL)

	log.Fatal(http.ListenAndServe(conf.ServerAddress, wh.HTTPRouter()))
}
