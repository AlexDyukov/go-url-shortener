package storage

import "strconv"

type User uint64

var DefaultUser = User(0)
var usersCount = uint64(1)

func NewUser() User {
	usersCount += 1
	return User(usersCount)
}

func ParseUser(userStr string) (User, error) {
	user, err := strconv.ParseUint(userStr, 10, 64)
	if err != nil {
		return DefaultUser, err
	}
	return User(user), nil
}

func UpdateUsersSeed(user User) {
	if usersCount < uint64(user) {
		usersCount = uint64(user)
	}
}
