package webconfig

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ServerAddress string

func (sa *ServerAddress) UnmarshalText(text []byte) error {
	return sa.Set(string(text))
}

func (sa *ServerAddress) String() string {
	return fmt.Sprint(*sa)
}

func (sa *ServerAddress) Set(value string) error {
	splitted := strings.Split(value, ":")
	if len(splitted) == 0 {
		return errors.New("should be \"address:port\"")
	}

	port := splitted[len(splitted)-1]
	if !isValidPort(port) {
		return errors.New("should be in range 1-65535")
	}

	hostname := strings.TrimSuffix(value, ":"+port)
	if !isValidHostname(hostname) {
		return errors.New("should be valid listen address")
	}

	*sa = ServerAddress(value)
	return nil
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	if p < 1 || p > 65535 {
		return false
	}
	return true
}

func isValidHostname(hostname string) bool {
	if hostname == "" || hostname == "localhost" {
		return true
	}
	//// IPv4
	if net.ParseIP(hostname) != nil {
		return true
	}
	//// IPv6
	if !strings.HasPrefix(hostname, "[") || !strings.HasSuffix(hostname, "]") {
		return false
	}
	serverIPv6Address := hostname[1 : len(hostname)-1]
	return net.ParseIP(serverIPv6Address) != nil
}
