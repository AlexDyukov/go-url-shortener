package urlshortener

import (
	"errors"
	"sync"

	"github.com/shomali11/util/xhashes"
)

var sMutex sync.RWMutex
var shorts map[uint64]string
var ErrConflict = errors.New("Conflict keys")

func init() {
	sMutex = sync.RWMutex{}
	shorts = map[uint64]string{}
}

func generateKey(value string) uint64 {
	return xhashes.FNV64a(value)
}

func GetLink(id uint64) (string, bool) {
	sMutex.RLock()
	link, exist := shorts[id]
	sMutex.RUnlock()

	if !exist {
		return "", false
	}

	return link, true
}

func MakeShort(newLink string) (uint64, error) {
	id := generateKey(newLink)

	sMutex.Lock()
	defer sMutex.Unlock()

	link, exist := shorts[id]
	if !exist {
		shorts[id] = newLink
		return id, nil
	}

	if link != newLink {
		return 0, ErrConflict
	}

	return id, nil
}
