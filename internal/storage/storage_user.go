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

type User int64

var DefaultUser = User(0)

func ParseUser(str []byte) (User, error) {
	str = bytes.TrimSpace(str)
	if len(str) == 0 {
		return DefaultUser, ErrInvalidUser{}
	}

	pos := 0
	u := int64(0)
	// sign
	sign := int64(1)
	if len(str) > 1 && str[pos] == '-' {
		sign = -sign
		pos += 1
	}
	// value
	for pos < len(str) && (str[pos] >= '0' && str[pos] <= '9') {
		if u > u*int64(10) { //overflow check
			return DefaultUser, ErrInvalidUser{}
		}
		u = u * int64(10)

		number := int64(str[pos] - '0')
		if u > u+number { //overflow check
			return DefaultUser, ErrInvalidUser{}
		}
		u = u + number

		pos += 1
	}
	if pos != len(str) {
		return DefaultUser, ErrInvalidUser{}
	}

	return User(sign * u), nil
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
