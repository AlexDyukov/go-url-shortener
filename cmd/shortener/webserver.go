package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	urlshortener "github.com/alexdyukov/go-url-shortener/internal/app"
	"github.com/julienschmidt/httprouter"
)

func listenAndServe(conf WebConfig) {
	router := httprouter.New()

	router.GET("/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id, err := strconv.ParseUint(ps.ByName("id"), 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link, exist := urlshortener.GetLink(id)
		if !exist {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(w, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		short, err := urlshortener.MakeShort(string(body))
		if err != nil {
			fmt.Fprintln(w, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)

		fmt.Fprintf(w, "http://localhost:8080/%d", short)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router)
}
