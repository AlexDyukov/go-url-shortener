package webhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/julienschmidt/httprouter"
)

type WebHandler struct {
	repo   service.Repository
	router *httprouter.Router
}

func NewWebHandler(svc service.Repository) *WebHandler {
	h := WebHandler{}
	h.repo = svc

	h.router = httprouter.New()
	h.router.GET("/:id", h.GetRoot)
	h.router.POST("/", h.PostRoot)
	h.router.POST("/api/shorten", h.PostAPIShorten)
	return &h
}

func (h *WebHandler) HTTPRouter() http.Handler {
	return h.router
}

func (h *WebHandler) GetRoot(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseUint(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, exist := h.repo.GetURL(storage.ID(id))
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *WebHandler) PostRoot(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := h.repo.SaveURL(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%v", id)
}

func (h *WebHandler) PostAPIShorten(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)

	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inputJSON := struct {
		URL string `json:"url"`
	}{}
	if err := json.Unmarshal(body, &inputJSON); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.repo.SaveURL(string(inputJSON.URL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outputJSON := struct {
		URL string `json:"result"`
	}{}
	outputJSON.URL = fmt.Sprintf("http://localhost:8080/%v", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(outputJSON)
}
