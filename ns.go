package main

import (
	"fmt"
	"net"
	"strings"
)

type NS struct {
	Server string `yaml:"server"`
	A      string `yaml:"a"`
	AAAA   string `yaml:"aaaa"`
}

func (ns *NS) checkError() error {
	if !strings.HasSuffix(ns.Server, ".catmunch") {
		return fmt.Errorf("nameserver domain name must end with .catmunch")
	}
	if ns.A != "" {
		ip := net.ParseIP(ns.A)
		if ip == nil {
			return fmt.Errorf("invalid ip")
		}
		if ip.To4() == nil {
			return fmt.Errorf("invalid ipv4")
		}
	}
	if ns.AAAA != "" {
		ip := net.ParseIP(ns.A)
		if ip == nil {
			return fmt.Errorf("invalid ip")
		}
		if ip.To4() != nil {
			return fmt.Errorf("invalid ipv6")
		}
	}
	return nil
}
