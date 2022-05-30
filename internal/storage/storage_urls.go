package storage

import (
	"strconv"

	"github.com/shomali11/util/xhashes"
)

type ErrConflict struct{}

func (e ErrConflict) Error() string {
	return "Storage: conflict keys"
}

type ErrInvalidShortID struct{}

func (e ErrInvalidShortID) Error() string {
	return "Storage: invalid shorted url"
}

type ErrNotFound struct{}

func (e ErrNotFound) Error() string {
	return "Storage: url not found"
}

type FullURL string
type CorrelationID string
type ShortID int64

var DefaultShortID = ShortID(0)
var DefaultFullURL = FullURL("")

func (urls SavedURLs) Save(sid ShortID, furl FullURL) error {
	savedurl, exist := urls[sid]
	if !exist {
		urls[sid] = furl
		return nil
	}

	if savedurl != furl {
		return ErrConflict{}
	}

	return nil
}

func (urls SavedURLs) Get(sid ShortID) (FullURL, error) {
	furl, exist := urls[sid]
	if !exist {
		return DefaultFullURL, ErrNotFound{}
	}
	return furl, nil
}

func short(furl FullURL) ShortID {
	return ShortID(xhashes.FNV64a(string(furl)))
}

func ParseShort(stringedShort string) (ShortID, error) {
	s, err := strconv.ParseInt(stringedShort, 10, 64)
	if err != nil {
		return DefaultShortID, ErrInvalidShortID{}
	}

	return ShortID(s), nil
}

func ParseCorrelationID(corrid string) CorrelationID {
	return CorrelationID(corrid)
}
