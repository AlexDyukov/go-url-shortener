package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	webhandler "github.com/alexdyukov/go-url-shortener/internal/webhandler"
)

func main() {
	conf := Config{}
	conf.ParseParams()

	stor := storage.NewInMemory()
	svc := service.NewURLShortener(stor)
	handler := webhandler.NewWebHandler(svc)

	router := httprouter.New()
	router.GET("/:id", handler.HandlerGet)
	router.POST("/", handler.HandlerPost)

	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router)
}
