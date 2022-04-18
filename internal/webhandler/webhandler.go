package webhandler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/julienschmidt/httprouter"
)

type WebHandler struct {
	service.Repository
}

func NewWebHandler(svc service.Repository) *WebHandler {
	return &WebHandler{svc}
}

func (h *WebHandler) HandlerGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseUint(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, exist := h.GetURL(storage.ID(id))
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *WebHandler) HandlerPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := h.SaveURL(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%v", id)
}
