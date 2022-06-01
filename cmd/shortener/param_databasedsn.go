package main

import (
	"fmt"

	pgx "github.com/jackc/pgx/v4"
)

type DataBaseDSN string

func (dbdsn *DataBaseDSN) UnmarshalText(text []byte) error {
	return dbdsn.Set(string(text))
}

func (dbdsn *DataBaseDSN) String() string {
	return fmt.Sprint(*dbdsn)
}

func (dbdsn *DataBaseDSN) Set(value string) error {
	if value == "" {
		return nil
	}
	_, err := pgx.ParseConfig(value)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	*dbdsn = DataBaseDSN(value)
	return nil
}
