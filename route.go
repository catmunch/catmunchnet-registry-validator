package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"strings"
)

type Route struct {
	CIDR        string   `yaml:"cidr"`
	Description string   `yaml:"description"`
	Origin      []string `yaml:"origin"`
}

var routes map[string]*Route

func LoadRoute(fileName string) (*Route, error) {
	if !InetRegex.MatchString(fileName) {
		return nil, fmt.Errorf("invalid route filename: %s", fileName)
	}
	var route Route
	bytes, err := os.ReadFile("route/" + fileName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &route)
	if err != nil {
		return nil, err
	}
	_, cidr, err := net.ParseCIDR(route.CIDR)
	if err != nil {
		return nil, err
	}
	if cidr.String() != route.CIDR {
		return nil, fmt.Errorf("incorrect cidr")
	}
	if cidr.IP.To4() == nil {
		return nil, fmt.Errorf("incorrect cidrv4")
	}
	if strings.ReplaceAll(route.CIDR, "/", "_") != fileName {
		return nil, fmt.Errorf("unmatching route '%s' is defined in route/%s", route.CIDR, fileName)
	}
	return &route, nil
}
func LoadRoutes() {
	routeFiles, err := os.ReadDir("route")
	panicErr(err)
	for _, routeFile := range routeFiles {
		fileName := routeFile.Name()
		if strings.HasPrefix(fileName, ".") {
			continue
		}
		route, err := LoadRoute(fileName)
		if err != nil {
			raiseError("invalid route content in upstream repo: " + fileName)
			panicErr(err)
		}
		routes[fileName] = route
	}
}
func ValidateRoutes() {
	for _, route := range routes {
		_, cidr, _ := net.ParseCIDR(route.CIDR)
		if err := IsValidRouteCIDR(cidr); err != nil {
			raiseError("invalid route CIDR " + route.CIDR + " : " + err.Error())
		}
		for _, asn := range route.Origin {
			if autnums[asn] == nil {
				raiseError("invalid route CIDR " + route.CIDR + " : ASN " + asn + " is not defined")
			}
		}
	}
}
