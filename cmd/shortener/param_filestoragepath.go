package main

import (
	"errors"
	"fmt"
	"os"
)

type FileStoragePath string

func (fsp *FileStoragePath) UnmarshalText(text []byte) error {
	return fsp.Set(string(text))
}

func (fsp *FileStoragePath) String() string {
	return fmt.Sprint(*fsp)
}

func (fsp *FileStoragePath) Set(value string) error {
	if value == "" {
		return nil
	}
	_, err := os.Stat(value)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("invalid file storage path: %w", err)
	}

	*fsp = FileStoragePath(value)
	return nil
}
