package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type Domain struct {
	Domain      string `yaml:"domain"`
	Description string `yaml:"description"`
	NS          []NS   `yaml:"ns"`
}

var DomainRegex = regexp.MustCompile(`^[a-zA-Z\d-_]+\.catmunch$`)
var domains map[string]*Domain

func LoadDomain(domainName string) (*Domain, error) {
	if !DomainRegex.MatchString(domainName) {
		return nil, fmt.Errorf("invalid domain: %s", domainName)
	}
	var domain Domain
	bytes, err := os.ReadFile("domain/" + domainName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &domain)
	if err != nil {
		return nil, err
	}
	if domain.Domain != domainName {
		return nil, fmt.Errorf("unmatching domain '%s' is defined in domain/%s", domain.Domain, domainName)
	}
	for _, ns := range domain.NS {
		if err := ns.checkError(); err != nil {
			return nil, err
		}
	}
	return &domain, nil
}
func LoadDomains() {
	domainFiles, err := os.ReadDir("domain")
	panicErr(err)
	for _, domainFile := range domainFiles {
		domainName := domainFile.Name()
		if strings.HasPrefix(domainName, ".") {
			continue
		}
		domain, err := LoadDomain(domainName)
		if err != nil {
			raiseError("invalid domain content in upstream repo: " + domainName)
			panicErr(err)
		}
		domains[domainName] = domain
	}
}
