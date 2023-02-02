package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"regexp"
	"strings"
)

type InetNum struct {
	CIDR        string `yaml:"cidr"`
	Description string `yaml:"description"`
	NS          []NS   `yaml:"ns"`
}

var InetRegex = regexp.MustCompile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})_(\d{1,2})$`)
var inetNums map[string]*InetNum

type TrieNode struct {
	Used  bool
	Dirty bool // used in subtree
	Child [2]*TrieNode
}

var cidrRoot TrieNode

func insertCIDR(net *net.IPNet) error {
	mask, _ := net.Mask.Size()
	if net.IP[0] != 10 || mask < 8 {
		return fmt.Errorf("CIDR must be in 10.0.0.0/8")
	}
	ptr := &cidrRoot
	mask -= 8
	bytePos := 1
	bitPos := 0
	for i := 0; i < mask; i++ {
		if ptr.Used {
			return fmt.Errorf("this CIDR is contained in a bigger ipv4 block")
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
		return fmt.Errorf("there is smaller ipv4 block inside this CIDR")
	}
	if ptr.Used {
		return fmt.Errorf("already defined this CIDR")
	}
	ptr.Used = true
	return nil
}
func IsValidRouteCIDR(net *net.IPNet) error {
	mask, _ := net.Mask.Size()
	if net.IP[0] != 10 || mask < 8 {
		return fmt.Errorf("CIDR must be in 10.0.0.0/8")
	}
	ptr := &cidrRoot
	mask -= 8
	bytePos := 1
	bitPos := 0
	for i := 0; i < mask; i++ {
		if ptr.Used {
			return nil
		}
		side := (net.IP[bytePos] >> (7 - bitPos)) & 1
		if ptr.Child[side] == nil {
			break
		}
		ptr = ptr.Child[side]
		bitPos++
		if bitPos == 8 {
			bytePos++
			bitPos = 0
		}
	}
	if ptr.Used {
		return nil
	}
	return fmt.Errorf("route is not contained in any defined IPv4 block")
}
func LoadInetNum(fileName string) (*InetNum, error) {
	if !InetRegex.MatchString(fileName) {
		return nil, fmt.Errorf("invalid inetnum filename: %s", fileName)
	}
	var inetNum InetNum
	bytes, err := os.ReadFile("inetnum/" + fileName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &inetNum)
	if err != nil {
		return nil, err
	}
	_, cidr, err := net.ParseCIDR(inetNum.CIDR)
	if err != nil {
		return nil, err
	}
	if cidr.String() != inetNum.CIDR {
		return nil, fmt.Errorf("incorrect cidr")
	}
	if cidr.IP.To4() == nil {
		return nil, fmt.Errorf("incorrect cidrv4")
	}
	if strings.ReplaceAll(inetNum.CIDR, "/", "_") != fileName {
		return nil, fmt.Errorf("unmatching inetnum '%s' is defined in inetnum/%s", inetNum.CIDR, fileName)
	}
	for _, ns := range inetNum.NS {
		if err := ns.checkError(); err != nil {
			return nil, err
		}
	}
	err = insertCIDR(cidr)
	if err != nil {
		return nil, err
	}
	return &inetNum, nil
}
func LoadInetNums() {
	inetNumFiles, err := os.ReadDir("inetnum")
	panicErr(err)
	for _, inetNumFile := range inetNumFiles {
		fileName := inetNumFile.Name()
		if strings.HasPrefix(fileName, ".") {
			continue
		}
		inetNum, err := LoadInetNum(fileName)
		if err != nil {
			raiseError("invalid inetnum content in upstream repo: " + fileName)
			panicErr(err)
		}
		inetNums[fileName] = inetNum
	}
}
