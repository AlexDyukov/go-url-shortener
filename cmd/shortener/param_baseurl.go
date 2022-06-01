package main

import (
	"fmt"
	"net/url"
)

type BaseURL string

func (burl *BaseURL) UnmarshalText(text []byte) error {
	return burl.Set(string(text))
}

func (burl *BaseURL) String() string {
	return fmt.Sprint(*burl)
}

func (burl *BaseURL) Set(value string) error {
	if _, err := url.ParseRequestURI(value); err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	*burl = BaseURL(value)
	return nil
}
