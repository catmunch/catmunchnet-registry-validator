package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"regexp"
	"strings"
)

type Inet6Num struct {
	CIDR        string `yaml:"cidr"`
	Description string `yaml:"description"`
	NS          []NS   `yaml:"ns"`
}

var Inet6Regex = regexp.MustCompile(`^([a-f0-9:]+)_(\d{1,2})$`)
var inet6Nums []*Inet6Num

var cidr6Root TrieNode

func insertCIDR6(net *net.IPNet) error {
	mask, _ := net.Mask.Size()
	if net.IP[0] != 0xfc || net.IP[1] != 0x75 || mask < 16 {
		return fmt.Errorf("CIDR must be in fc75::/16")
	}
	ptr := &cidrRoot
	mask -= 16
	bytePos := 2
	bitPos := 0
	for i := 0; i < mask; i++ {
		if ptr.Used {
			return fmt.Errorf("this CIDR is contained in a bigger ipv6 block")
		}
		ptr.Dirty = true
		side := (net.IP[bytePos] >> (7 - bitPos)) & 1
		if ptr.Child[side] == nil {
			ptr.Child[side] = &TrieNode{}
		}
		ptr = ptr.Child[side]
		bitPos++
		if bitPos == 8 {
			bytePos++
			bitPos = 0
		}
	}
	if ptr.Dirty {
		return fmt.Errorf("there is smaller ipv6 block inside this CIDR")
	}
	if ptr.Used {
		return fmt.Errorf("already defined this CIDR")
	}
	ptr.Used = true
	return nil
}
func LoadInet6Num(fileName string) (*Inet6Num, error) {
	if !Inet6Regex.MatchString(fileName) {
		panicErr(fmt.Errorf("invalid inet6num filename: %s", fileName))
	}
	var inet6Num Inet6Num
	bytes, err := os.ReadFile("inet6num/" + fileName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &inet6Num)
	if err != nil {
		return nil, err
	}
	_, cidr, err := net.ParseCIDR(inet6Num.CIDR)
	if err != nil {
		return nil, err
	}
	if cidr.String() != inet6Num.CIDR {
		return nil, fmt.Errorf("incorrect cidr")
	}
	if cidr.IP.To4() != nil {
		return nil, fmt.Errorf("incorrect cidrv6")
	}
	if strings.ReplaceAll(inet6Num.CIDR, "/", "_") != fileName {
		return nil, fmt.Errorf("unmatching inet6num '%s' is defined in inet6num/%s", inet6Num.CIDR, fileName)
	}
	for _, ns := range inet6Num.NS {
		if err := ns.checkError(); err != nil {
			return nil, err
		}
	}
	err = insertCIDR6(cidr)
	if err != nil {
		return nil, err
	}
	return &inet6Num, nil
}
func LoadInet6Nums() {
	inet6NumFiles, err := os.ReadDir("inet6num")
	panicErr(err)
	for _, inet6NumFile := range inet6NumFiles {
		fileName := inet6NumFile.Name()
		if strings.HasPrefix(fileName, ".") {
			continue
		}
		inet6Num, err := LoadInet6Num(fileName)
		if err != nil {
			raiseError("invalid inet6num content in upstream repo: " + fileName)
			panicErr(err)
		}
		inet6Nums = append(inet6Nums, inet6Num)
	}
}
