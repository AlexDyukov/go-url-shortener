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

func (u *URLShortener) getCorrelationID(corrid storage.CorrelationID) string {
	return fmt.Sprint(corrid)
}

func (u *URLShortener) SaveURL(ctx context.Context, fullURL string) (string, error) {
	if !isValidURL(fullURL) {
		return "", ErrInvalidURL{}
	}

	furl := storage.FullURL(fullURL)

	sid, err := u.stor.Put(ctx, furl)
	if err != nil {
		return u.getShortURL(sid), err
	}

	return u.getShortURL(sid), nil
}

func (u *URLShortener) SaveBatch(ctx context.Context, breq []BatchRequestItem) ([]BatchResponseItem, error) {
	storRequest := storage.BatchRequest{}
	for _, v := range breq {
		corrid := storage.ParseCorrelationID(v.CorrelationID)
		furl := storage.FullURL(v.OriginalURL)
		storRequest[corrid] = furl
	}

	storResponse, err := u.stor.PutBatch(ctx, storRequest)
	result := []BatchResponseItem{}
	for corrid, sid := range storResponse {
		result = append(result, BatchResponseItem{CorrelationID: u.getCorrelationID(corrid), ShortURL: u.getShortURL(sid)})
	}
	return result, err
}

func (u *URLShortener) GetURL(ctx context.Context, shortIDstr string) (string, error) {
	sid, err := storage.ParseShort(shortIDstr)
	if err != nil {
		return "", err
	}

	furl, err := u.stor.Get(ctx, sid)
	if err != nil {
		return "", err
	}
	return u.getFullURL(furl), nil
}

func (u *URLShortener) GetURLs(ctx context.Context) ([]URLs, error) {
	urls, err := u.stor.GetURLs(ctx)
	if err != nil {
		return []URLs{}, err
	}

	answer := []URLs{}
	for sid, furl := range urls {
		answer = append(answer, URLs{Short: u.getShortURL(sid), Original: u.getFullURL(furl)})
	}

	return answer, nil
}

func (u *URLShortener) DeleteURLs(ctx context.Context, todelete []string) error {
	sids := []storage.ShortID{}
	for _, shortIDstr := range todelete {
		sid, err := storage.ParseShort(shortIDstr)
		if err != nil {
			return err
		}
		sids = append(sids, sid)
	}

	return u.stor.DeleteURLs(ctx, sids)
}

func (u *URLShortener) NewUser(ctx context.Context) (storage.User, error) {
	return u.stor.NewUser(ctx)
}

func (u *URLShortener) Ping(ctx context.Context) bool {
	return u.stor.Ping(ctx)
}
