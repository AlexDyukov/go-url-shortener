package service

import (
	"context"
	"fmt"
	"strconv"

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

func (u *URLShortener) SaveURL(ctx context.Context, url string) (string, error) {
	if !isValidURL(url) {
		return "", ErrInvalidURL{}
	}

	user := getUser(ctx)
	furl := storage.FullURL(url)

	sid, err := u.stor.Put(user, furl)
	if err != nil {
		return "", err
	}

	return u.getShortURL(sid), nil
}

func (u *URLShortener) GetURL(ctx context.Context, shortURLstr string) (string, bool) {
	parsedShortID, err := strconv.ParseUint(shortURLstr, 10, 64)
	if err != nil {
		return "", false
	}

	sid := storage.ShortID(parsedShortID)
	user := getUser(ctx)
	furl, exist := u.stor.Get(user, sid)
	if !exist {
		return "", false
	}
	return u.getFullURL(furl), true
}

func (u *URLShortener) GetURLs(ctx context.Context) []URLs {
	user := getUser(ctx)

	answer := []URLs{}

	if user == storage.DefaultUser {
		return answer
	}

	for sid, furl := range u.stor.GetURLs(user) {
		answer = append(answer, URLs{Short: u.getShortURL(sid), Original: u.getFullURL(furl)})
	}

	return answer
}
