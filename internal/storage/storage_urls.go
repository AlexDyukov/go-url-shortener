package storage

import "github.com/shomali11/util/xhashes"

type FullURL string
type ShortID uint64
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

func (urls URLs) Get(sid ShortID) (FullURL, bool) {
	furl, exist := urls[sid]
	return furl, exist
}

func short(furl FullURL) ShortID {
	return ShortID(xhashes.FNV64a(string(furl)))
}
