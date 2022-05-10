package main

import (
	"errors"
	"flag"
	"fmt"
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
}

func (c *Config) Parse() {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&c.ServerAddress, "address", c.ServerAddress, "http listen address in \"address:port\" format")
	flag.StringVar(&c.BaseURL, "baseurl", c.BaseURL, "base url for shortener")
	flag.StringVar(&c.FileStoragePath, "file-storage-path", c.FileStoragePath, "base url for shortener")
	flag.Parse()

	if !c.isValidServerAddress() {
		logStr := fmt.Sprintf("invalid value \"%s\" for address\n", c.ServerAddress)
		log.Fatal(logStr)
	}

	if !c.isValidBaseURL() {
		logStr := fmt.Sprintf("invalid value \"%s\" for base URL\n", c.BaseURL)
		log.Fatal(logStr)
	}

	if !c.isValidFileStoragePath() {
		logStr := fmt.Sprintf("invalid value \"%s\" for file storage path\n", c.FileStoragePath)
		log.Fatal(logStr)
	}
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

	// validate server address
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
