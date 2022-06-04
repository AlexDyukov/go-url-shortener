package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type ErrInvalidUser struct{}

func (e ErrInvalidUser) Error() string {
	return "Storage: invalid user"
}

type User int64
type UserCtxKey struct{}

var DefaultUser = User(0)

func ParseUser(stringedUser string) (User, error) {
	stringedUser = strings.TrimSpace(stringedUser)
	u, err := strconv.ParseInt(stringedUser, 10, 64)
	if err != nil {
		return DefaultUser, ErrInvalidUser{}
	}
	return User(u), nil
}

func GetUser(ctx context.Context) (User, error) {
	user, ok := ctx.Value(UserCtxKey{}).(string)
	if !ok {
		return DefaultUser, ErrInvalidUser{}
	}

	return ParseUser(user)
}

func PutUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, fmt.Sprint(user))
}
