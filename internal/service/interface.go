package service

import storage "github.com/alexdyukov/go-url-shortener/internal/storage"

type ErrInvalidURL struct{}

func (e ErrInvalidURL) Error() string {
	return "Repository: invalid URL"
}

type Repository interface {
	SaveURL(url string) (storage.ID, error)
	GetURL(id storage.ID) (string, bool)
}
