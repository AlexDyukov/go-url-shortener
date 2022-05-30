package service

import (
	"context"

	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

type ErrInvalidURL struct{}

func (e ErrInvalidURL) Error() string {
	return "Repository: invalid URL"
}

type URLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Repository interface {
	SaveURL(ctx context.Context, fullURL string) (string, error)
	SaveBatch(ctx context.Context, breq []BatchRequestItem) ([]BatchResponseItem, error)
	GetURL(ctx context.Context, shortIDstr string) (string, error)
	GetURLs(ctx context.Context) ([]URLs, error)
	NewUser(ctx context.Context) (storage.User, error)
	Ping(ctx context.Context) bool
}

func isValidURL(url string) bool {
	return url != ""
}
