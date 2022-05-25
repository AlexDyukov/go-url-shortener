package storage

import "github.com/shomali11/util/xhashes"

type ErrConflict struct{}

func (e ErrConflict) Error() string {
	return "Storage: conflict keys"
}

type FullURL string
type ShortURL uint64
type URLs map[ShortURL]FullURL

type User uint64
type UsersShorts map[User]URLs

// backward compatibility
var DefaultUserID = User(0)

func (urls URLs) Save(surl ShortURL, furl FullURL) error {
	savedurl, exist := urls[surl]
	if !exist {
		urls[surl] = furl
		return nil
	}

	if savedurl != furl {
		return ErrConflict{}
	}

	return nil
}

func (urls URLs) Get(surl ShortURL) (FullURL, bool) {
	furl, exist := urls[surl]
	return furl, exist
}

type Storage interface {
	Get(surl ShortURL) (FullURL, bool)
	Save(surl ShortURL, furl FullURL) error
	Put(furl FullURL) (ShortURL, error)
	SetAuthor(surl ShortURL, furl FullURL, user User) error
	GetAuthorURLs(user User) (URLs, bool)
}

func short(furl FullURL) ShortURL {
	return ShortURL(xhashes.FNV64a(string(furl)))
}
