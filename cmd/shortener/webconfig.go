package main

import (
	"flag"
	"log"
	"sync"

	env "github.com/caarlos0/env/v6"
)

type Config struct {
	once            sync.Once
	ServerAddress   ServerAddress   `env:"SERVER_ADDRESS" envDefault:":8080" envExpand:"true"`
	BaseURL         BaseURL         `env:"BASE_URL" envDefault:"http://localhost:8080" envExpand:"true"`
	FileStoragePath FileStoragePath `env:"FILE_STORAGE_PATH" envDefault:"" envExpand:"true"`
	EncryptKey      EncryptKey      `env:"ENCRYPT_KEY" envDefault:"testtesttesttest" envExpand:"true"`
	DataBaseDSN     DataBaseDSN     `env:"DATABASE_DSN" envDefault:"" envExpand:"true"`
}

var config Config

func init() {
	flag.Var(&config.ServerAddress, "a", "http listen address ")
	flag.Var(&config.BaseURL, "b", "base url for shortener")
	flag.Var(&config.FileStoragePath, "f", "path to storage file")
	flag.Var(&config.EncryptKey, "k", "16 bit encrypt key for auth cookie")
	flag.Var(&config.DataBaseDSN, "d", "database DSN link")
}

func GetConfig() *Config {
	config.once.Do(func() {
		if err := env.Parse(&config); err != nil {
			log.Fatal(err)
		}
		flag.Parse()
	})

	return &config
}
