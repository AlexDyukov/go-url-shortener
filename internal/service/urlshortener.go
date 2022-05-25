package service

import (
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

type URLShortener struct {
	storage.Storage
}

func isValid(url string) bool {
	return url != ""
}

func (u *URLShortener) SaveURL(url string) (storage.ShortURL, error) {
	if !isValid(url) {
		return storage.ShortURL(0), ErrInvalidURL{}
	}

	return u.Put(storage.FullURL(url))
}

func (u *URLShortener) GetURL(short storage.ShortURL) (string, bool) {
	furl, exist := u.Get(short)
	return string(furl), exist
}

func NewURLShortener(s storage.Storage) Repository {
	return &URLShortener{s}
}
