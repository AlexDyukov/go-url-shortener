package storage

import (
	"bytes"

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
type ShortID int64

var DefaultShortID = ShortID(0)
var DefaultFullURL = FullURL("")

type URLs map[ShortID]FullURL

func (urls URLs) Save(sid ShortID, furl FullURL) error {
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

func (urls URLs) Get(sid ShortID) (FullURL, error) {
	furl, exist := urls[sid]
	if !exist {
		return DefaultFullURL, ErrNotFound{}
	}
	return furl, nil
}

func short(furl FullURL) ShortID {
	sid := ShortID(xhashes.FNV64a(string(furl)))
	if sid < 0 {
		return -sid
	}
	return sid
}

func ParseShort(str []byte) (ShortID, error) {
	str = bytes.TrimSpace(str)
	if len(str) == 0 {
		return DefaultShortID, ErrInvalidShortID{}
	}

	pos := 0
	shorted := int64(0)
	for pos < len(str) && (str[pos] >= '0' && str[pos] <= '9') {
		if shorted > shorted*int64(10) { //overflow check
			return DefaultShortID, ErrInvalidShortID{}
		}
		shorted = shorted * int64(10)

		number := int64(str[pos] - '0')
		if shorted > shorted+number { //overflow check
			return DefaultShortID, ErrInvalidShortID{}
		}
		shorted = shorted + number

		pos += 1
	}
	if pos != len(str) {
		return DefaultShortID, ErrInvalidShortID{}
	}

	return ShortID(shorted), nil
}
