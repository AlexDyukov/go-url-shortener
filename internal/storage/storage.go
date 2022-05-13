package storage

import "github.com/shomali11/util/xhashes"

type ErrConflict struct{}

func (e ErrConflict) Error() string {
	return "Storage: conflict keys"
}

type ID uint64

type Storage interface {
	Get(id ID) (string, bool)
	Put(str string) (ID, error)
	Save(id ID, link string)
}

func hash(str string) ID {
	return ID(xhashes.FNV64a(str))
}
