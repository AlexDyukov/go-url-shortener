package main

import (
	"flag"
	"fmt"
	"log"
)

type WebConfig struct {
	Port int
}

func (c *WebConfig) ParseParams() {
	flag.IntVar(&c.Port, "port", 8080, "http listen port, 1025-65535")
	flag.Parse()

	if c.Port < 1025 || c.Port > 65535 {
		logStr := fmt.Sprintf("invalid value \"%d\" for flag -port: should be in range [1025;65535]\n", c.Port)
		log.Fatal(logStr)
	}
}
