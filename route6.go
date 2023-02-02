package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"strings"
)

type Route6 struct {
	CIDR        string   `yaml:"cidr"`
	Description string   `yaml:"description"`
	Origin      []string `yaml:"origin"`
}

var routes6 map[string]*Route6

func LoadRoute6(fileName string) (*Route6, error) {
	if !InetRegex.MatchString(fileName) {
		return nil, fmt.Errorf("invalid route6 filename: %s", fileName)
	}
	var route Route6
	bytes, err := os.ReadFile("route6/" + fileName)
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
	if cidr.IP.To4() != nil {
		return nil, fmt.Errorf("incorrect cidrv6")
	}
	if strings.ReplaceAll(route.CIDR, "/", "_") != fileName {
		return nil, fmt.Errorf("unmatching route6 '%s' is defined in route6/%s", route.CIDR, fileName)
	}
	return &route, nil
}
func LoadRoutes6() {
	routeFiles, err := os.ReadDir("route6")
	panicErr(err)
	for _, routeFile := range routeFiles {
		fileName := routeFile.Name()
		if strings.HasPrefix(fileName, ".") {
			continue
		}
		route, err := LoadRoute6(fileName)
		if err != nil {
			raiseError("invalid route6 content in upstream repo: " + fileName)
			panicErr(err)
		}
		routes6[fileName] = route
	}
}
func ValidateRoutes6() {
	for _, route := range routes6 {
		_, cidr, _ := net.ParseCIDR(route.CIDR)
		if err := IsValidRouteCIDR6(cidr); err != nil {
			raiseError("invalid route6 CIDR " + route.CIDR + " : " + err.Error())
		}
		for _, asn := range route.Origin {
			if autnums[asn] == nil {
				raiseError("invalid route6 CIDR " + route.CIDR + " : ASN " + asn + " is not defined")
			}
		}
	}
}
func GenerateROA6() string {
	result := ""
	for _, route := range routes6 {
		_, cidr, _ := net.ParseCIDR(route.CIDR)
		mask, _ := cidr.Mask.Size()
		for _, asn := range route.Origin {
			result += fmt.Sprintf("route %s max %d as %s;\n", route.CIDR, mask, asn)
		}
	}
	return result
}
