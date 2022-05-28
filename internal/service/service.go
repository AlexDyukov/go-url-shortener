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

type Repository interface {
	SaveURL(ctx context.Context, fullURL string) (string, error)
	GetURL(ctx context.Context, shortIDstr string) (string, bool)
	GetURLs(ctx context.Context) []URLs
	NewUser(ctx context.Context) storage.User
	Ping(ctx context.Context) bool
}

func isValidURL(url string) bool {
	return url != ""
}
