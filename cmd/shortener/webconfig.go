package main

import (
	"flag"
	"fmt"
	"log"
)

type Config struct {
	Port int
}

func (c *Config) ParseParams() {
	flag.IntVar(&c.Port, "port", 8080, "http listen port, 1025-65535")
	flag.Parse()

	if c.Port < 1025 || c.Port > 65535 {
		logStr := fmt.Sprintf("invalid value \"%d\" for flag -port: should be in range [1025;65535]\n", c.Port)
		log.Fatal(logStr)
	}
}
