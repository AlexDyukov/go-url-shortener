package webhandler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/gorilla/mux"
)

type WebHandler struct {
	repo      service.Repository
	router    *mux.Router
	encryptor *Encryptor
}

func NewWebHandler(svc service.Repository, encryptKey string) *WebHandler {
	h := WebHandler{}
	h.repo = svc
	h.encryptor = newEncryptor([]byte(encryptKey))

	router := mux.NewRouter()
	router.HandleFunc("/{id:[-]?[0-9]+}", h.GetRoot).Methods("GET")
	router.HandleFunc("/", h.PostRoot).Methods("POST")
	router.HandleFunc("/api/shorten", h.PostAPIShorten).Methods("POST")
	router.HandleFunc("/api/shorten/batch", h.PostAPIShortenBatch).Methods("POST")
	router.HandleFunc("/api/user/urls", h.GetAPIUserURLs).Methods("GET")
	router.HandleFunc("/api/user/urls", h.DeleteAPIUserURLs).Methods("DELETE")
	router.HandleFunc("/ping", h.Ping).Methods("GET")

	h.router = router

	return &h
}

func (h *WebHandler) HTTPRouter() http.Handler {
	ah := newAuthHandler(h.encryptor, h.repo)
	handler := ah(h.router)
	handler = compressHandler(handler)
	return handler
}

func (h *WebHandler) GetRoot(w http.ResponseWriter, r *http.Request) {
	url, err := h.repo.GetURL(r.Context(), mux.Vars(r)["id"])

	switch err.(type) {
	case nil:
	case storage.ErrInvalidShortID:
		w.WriteHeader(http.StatusBadRequest)
		return
	case storage.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
		return
	case storage.ErrDeleted:
		w.WriteHeader(http.StatusGone)
		return
	default:
		log.Println("webhandler: GetRoot: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *WebHandler) PostRoot(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("webhandler: PostRoot: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL, err := h.repo.SaveURL(r.Context(), string(body))
	switch err.(type) {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case storage.ErrConflict:
		w.WriteHeader(http.StatusConflict)
	case service.ErrInvalidURL:
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		log.Println("webhandler: PostRoot: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, shortURL)
}

func (h *WebHandler) PostAPIShorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("webhandler: PostAPIShorten: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	inputJSON := struct {
		URL string `json:"url"`
	}{}
	if err := json.Unmarshal(body, &inputJSON); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.repo.SaveURL(r.Context(), string(inputJSON.URL))
	switch err.(type) {
	case nil:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	case storage.ErrConflict:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	case service.ErrInvalidURL:
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		log.Println("webhandler: PostAPIShorten: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outputJSON := struct {
		URL string `json:"result"`
	}{}
	outputJSON.URL = shortURL
	json.NewEncoder(w).Encode(outputJSON)
}

func (h *WebHandler) PostAPIShortenBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("webhandler: PostAPIShortenBatch: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	input := []service.BatchRequestItem{}
	if err := json.Unmarshal(body, &input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := h.repo.SaveBatch(r.Context(), input)
	if err != nil {
		log.Println("webhandler: PostAPIShortenBatch: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

func (h *WebHandler) GetAPIUserURLs(w http.ResponseWriter, r *http.Request) {
	urls, err := h.repo.GetURLs(r.Context())
	switch err.(type) {
	case nil:
	case storage.ErrNotFound:
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		log.Println("webhandler: GetAPIUserURLs: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	encoder.Encode(urls)
}

func (h *WebHandler) DeleteAPIUserURLs(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("webhandler: DeleteAPIUserURLs: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	input := []string{}
	if err := json.Unmarshal(body, &input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteURLs(r.Context(), input)
	switch err.(type) {
	case nil:
	case storage.ErrNotFound:
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		log.Println("webhandler: GetAPIUserURLs: InternalServerError:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *WebHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if !h.repo.Ping(r.Context()) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
