package service

import (
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

type URLShortener struct {
	storage.Storage
}

func isValid(url string) bool {
	return true
}

func (u *URLShortener) SaveURL(url string) (storage.ID, error) {
	if !isValid(url) {
		return storage.ID(0), ErrInvalidURL{}
	}

	return u.Put(url)
}

func (u *URLShortener) GetURL(id storage.ID) (string, bool) {
	return u.Get(id)
}

func NewURLShortener(s storage.Storage) Repository {
	return &URLShortener{s}
}
