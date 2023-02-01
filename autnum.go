package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Autnum struct {
	Autnum      string `yaml:"autnum"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var ASNRegex = regexp.MustCompile(`^AS(\d+)$`)
var autnums []*Autnum

func LoadAutnum(asn string) (*Autnum, error) {
	if !ASNRegex.MatchString(asn) {
		return nil, fmt.Errorf("invalid asn: %s", asn)
	}
	asnDigit, err := strconv.ParseInt(ASNRegex.FindStringSubmatch(asn)[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid asn: %s", asn)
	}
	if asnDigit < 64512 || (asnDigit > 65534 && asnDigit < 4200000000) || asnDigit >= 4294967295 {
		return nil, fmt.Errorf("asn is for public use: %s", asn)
	}
	var autnum Autnum
	bytes, err := os.ReadFile("autnum/" + asn)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &autnum)
	if err != nil {
		return nil, err
	}
	if autnum.Autnum != asn {
		return nil, fmt.Errorf("unmatching AS number '%s' is defined in autnum/%s", autnum.Autnum, asn)
	}
	if autnum.Name == "" {
		return nil, fmt.Errorf("name is missing in autnum/%s", asn)
	}
	return &autnum, nil
}
func LoadAutnums() {
	autnumFiles, err := os.ReadDir("autnum")
	panicErr(err)
	for _, autnumFile := range autnumFiles {
		asn := autnumFile.Name()
		if strings.HasPrefix(asn, ".") {
			continue
		}
		autnum, err := LoadAutnum(asn)
		if err != nil {
			raiseError("invalid asn content in upstream repo: " + asn)
			panicErr(err)
		}
		autnums = append(autnums, autnum)
	}
}
