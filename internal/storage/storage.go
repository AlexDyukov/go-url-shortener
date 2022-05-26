package storage

type ErrConflict struct{}

func (e ErrConflict) Error() string {
	return "Storage: conflict keys"
}

type Storage interface {
	Get(user User, sid ShortID) (FullURL, bool)
	Save(user User, sid ShortID, furl FullURL) error
	Put(user User, furl FullURL) (ShortID, error)
	GetURLs(user User) URLs
}
