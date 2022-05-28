package service

import (
	"context"
	"fmt"

	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

type URLShortener struct {
	stor    storage.Storage
	baseURL string
}

func NewURLShortener(s storage.Storage, baseURL string) Repository {
	return &URLShortener{s, baseURL}
}

func (u *URLShortener) getShortURL(sid storage.ShortID) string {
	return fmt.Sprintf("%s/%v", u.baseURL, sid)
}

func (u *URLShortener) getFullURL(furl storage.FullURL) string {
	return fmt.Sprint(furl)
}

func (u *URLShortener) SaveURL(ctx context.Context, fullURL string) (string, error) {
	if !isValidURL(fullURL) {
		return "", ErrInvalidURL{}
	}

	furl := storage.FullURL(fullURL)

	sid, err := u.stor.Put(ctx, furl)
	if err != nil {
		return "", err
	}

	return u.getShortURL(sid), nil
}

func (u *URLShortener) GetURL(ctx context.Context, shortIDstr string) (string, bool) {
	sid, err := storage.ParseShort([]byte(shortIDstr))
	if err != nil {
		return "", false
	}

	furl, exist := u.stor.Get(ctx, sid)
	if !exist {
		return "", false
	}
	return u.getFullURL(furl), true
}

func (u *URLShortener) GetURLs(ctx context.Context) []URLs {
	answer := []URLs{}
	for sid, furl := range u.stor.GetURLs(ctx) {
		answer = append(answer, URLs{Short: u.getShortURL(sid), Original: u.getFullURL(furl)})
	}

	return answer
}

func (u *URLShortener) NewUser(ctx context.Context) storage.User {
	return u.stor.NewUser(ctx)
}

func (u *URLShortener) Ping(ctx context.Context) bool {
	return u.stor.Ping(ctx)
}
