package webconfig

import (
	"errors"
	"fmt"
)

type EncryptKey string

func (ek *EncryptKey) UnmarshalText(text []byte) error {
	return ek.Set(string(text))
}

func (ek *EncryptKey) String() string {
	return fmt.Sprint(*ek)
}

func (ek *EncryptKey) Set(value string) error {
	if len(value) != 16 {
		return errors.New("invalid encrypt key")
	}

	*ek = EncryptKey(value)
	return nil
}
