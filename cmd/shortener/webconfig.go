package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	env "github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080" envExpand:"true"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080" envExpand:"true"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"" envExpand:"true"`
	EncryptKey      string `env:"ENCRYPT_KEY" envDefault:"testtesttesttest" envExpand:"true"`
}

var config = Config{}

func init() {
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&config.ServerAddress, "a", config.ServerAddress, "http listen address in \"address:port\" format")
	flag.StringVar(&config.BaseURL, "b", config.BaseURL, "base url for shortener")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "path to storage file")
	flag.StringVar(&config.EncryptKey, "k", config.EncryptKey, "16 bit encrypt key for auth cookie")
	flag.Parse()

	if !config.isValidServerAddress() {
		log.Fatal("invalid value for address:", config.ServerAddress)
	}

	if !config.isValidBaseURL() {
		log.Fatal("invalid value for base URL:", config.BaseURL)
	}

	if !config.isValidFileStoragePath() {
		log.Fatal("invalid value for file storage path:", config.FileStoragePath)
	}

	if !config.isValidEncryptKey() {
		log.Fatal("invalid value for encrypt key:", config.EncryptKey)
	}
}

func GetConfig() *Config {
	return &config
}

func (c *Config) isValidEncryptKey() bool {
	return true
}

func (c *Config) isValidFileStoragePath() bool {
	_, err := os.Stat(c.FileStoragePath)
	return err == nil || errors.Is(err, os.ErrNotExist)
}

func (c *Config) isValidBaseURL() bool {
	_, err := url.ParseRequestURI(c.BaseURL)
	return err == nil
}

func (c *Config) isValidServerAddress() bool {
	splitted := strings.Split(c.ServerAddress, ":")
	if len(splitted) == 0 {
		return false
	}

	// validate port
	portStr := splitted[len(splitted)-1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	if port < 1 || port > 65535 {
		return false
	}

	// validate host
	hostStr := strings.TrimSuffix(c.ServerAddress, ":"+portStr)
	if hostStr == "" || hostStr == "localhost" {
		return true
	}
	//// IPv4
	if net.ParseIP(hostStr) != nil {
		return true
	}
	//// IPv6
	if !strings.HasPrefix(hostStr, "[") || !strings.HasSuffix(hostStr, "]") {
		return false
	}
	serverIPv6Address := hostStr[1 : len(hostStr)-1]
	return net.ParseIP(serverIPv6Address) != nil
}
