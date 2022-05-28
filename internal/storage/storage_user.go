package storage

import (
	"bytes"
	"context"
	"fmt"
)

type ErrInvalidUser struct{}

func (e ErrInvalidUser) Error() string {
	return "Storage: invalid user"
}

type User uint64

var DefaultUser = User(0)

func ParseUser(str []byte) (User, error) {
	str = bytes.TrimSpace(str)
	if len(str) == 0 {
		return DefaultUser, ErrInvalidUser{}
	}

	pos := 0
	u := uint64(0)
	for pos < len(str) && (str[pos] >= '0' && str[pos] <= '9') {
		if u > u*uint64(10) { //overflow check
			return DefaultUser, ErrInvalidUser{}
		}
		u = u * uint64(10)

		number := uint64(str[pos] - '0')
		if u > u+number { //overflow check
			return DefaultUser, ErrInvalidUser{}
		}
		u = u + number

		pos += 1
	}
	if pos != len(str) {
		return DefaultUser, ErrInvalidUser{}
	}

	return User(u), nil
}

func GetUser(ctx context.Context) (User, error) {
	input, ok := ctx.Value(UserCtxKey{}).(string)
	if !ok {
		return DefaultUser, ErrInvalidUser{}
	}

	return ParseUser([]byte(input))
}

func PutUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, fmt.Sprint(user))
}
