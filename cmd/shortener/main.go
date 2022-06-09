package main

import (
	"log"
	"net/http"

	"github.com/alexdyukov/go-url-shortener/internal/service"
	"github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/alexdyukov/go-url-shortener/internal/webhandler"
	"github.com/alexdyukov/go-url-shortener/cmd/webconfig"
)

func main() {
	conf := webconfig.GetConfig()

	var stor storage.Storage
	switch {
	case conf.DataBaseDSN != "":
		s, err := storage.NewInDatabase(conf.DataBaseDSN.String())
		if err != nil {
			log.Fatal("cannot open database connection:", err.Error())
		}
		stor = s
	case conf.FileStoragePath != "":
		s, err := storage.NewInFile(conf.FileStoragePath.String())
		if err != nil {
			log.Fatal("cannot open storage file:", err.Error())
		}
		stor = s
	default:
		stor = storage.NewInMemory()
	}
	svc := service.NewURLShortener(stor, conf.BaseURL.String())
	wh := webhandler.NewWebHandler(svc, conf.EncryptKey.String())

	log.Fatal(http.ListenAndServe(conf.ServerAddress.String(), wh.HTTPRouter()))
}
