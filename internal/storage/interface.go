package storage

type ErrConflict struct{}

func (e ErrConflict) Error() string {
	return "Storage: conflict keys"
}

type ID uint64

type Storage interface {
	Get(id ID) (string, bool)
	Put(str string) (ID, error)
}
