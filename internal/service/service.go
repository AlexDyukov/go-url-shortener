package service

import (
	"context"

	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

type ErrInvalidURL struct{}

func (e ErrInvalidURL) Error() string {
	return "Repository: invalid URL"
}

type UserCtxKey struct{}

type URLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

type Repository interface {
	SaveURL(ctx context.Context, fullURL string) (string, error)
	GetURL(ctx context.Context, shortURLID string) (string, bool)
	GetURLs(ctx context.Context) []URLs
}

func getUser(ctx context.Context) storage.User {
	if value, ok := ctx.Value(UserCtxKey{}).(storage.User); ok {
		return value
	}
	return storage.DefaultUser
}

func setUser(ctx context.Context, user storage.User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, user)
}

func isValidURL(url string) bool {
	return url != ""
}
