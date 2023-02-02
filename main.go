package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"sort"
	"strings"
)

type Context struct {
	Repository     *git.Repository
	PrCommit       *object.Commit
	UpstreamCommit *object.Commit
	Patch          *object.Patch
}

var valid = true

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
func getCommit(context *Context) {
	workdir, err := os.Getwd()
	panicErr(err)
	repo, err := git.PlainOpen(workdir)
	panicErr(err)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{os.Getenv("CI_REPOSITORY_URL")},
		Fetch: []config.RefSpec{
			"+refs/heads/main:refs/remotes/upstream/main",
		},
	})
	panicErr(err)
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "upstream",
	})
	panicErr(err)
	prHash, err := repo.ResolveRevision("HEAD")
	panicErr(err)
	upstreamHash, err := repo.ResolveRevision("refs/remotes/upstream/main")
	panicErr(err)
	context.Repository = repo
	context.PrCommit, err = repo.CommitObject(*prHash)
	panicErr(err)
	context.UpstreamCommit, err = repo.CommitObject(*upstreamHash)
	panicErr(err)
	context.Patch, err = context.UpstreamCommit.Patch(context.PrCommit)
	return
}

func raiseError(s string) {
	_, _ = fmt.Fprintln(os.Stderr, s)
	valid = false
}
func checkChangedResources(context *Context) {
	patches := context.Patch.FilePatches()
	sort.Slice(patches, func(i, j int) bool {
		from1, to1 := patches[i].Files()
		from2, to2 := patches[i].Files()
		if to1 == nil {
			return true
		}
		if to2 == nil {
			return false
		}
		if from1 == nil {
			return true
		}
		if from2 == nil {
			return false
		}
		return false
	})
	for _, patch := range patches {
		from, to := patch.Files()
		resourcePath := ""
		if from == nil {
			resourcePath = to.Path()
		} else {
			resourcePath = from.Path()
		}
		if strings.HasPrefix(resourcePath, "autnum/") {
			asn := strings.TrimPrefix(resourcePath, "autnum/")
			if from == nil {
				autnum, err := LoadAutnum(asn)
				if err != nil {
					raiseError("autnum/" + asn + " is invalid: " + err.Error())
				} else {
					fmt.Println("New ASN: " + autnum.Autnum + "(" + autnum.Name + ")")
					autnums[asn] = autnum
				}
			} else if to == nil {
				fmt.Println("Deleted ASN: " + autnums[asn].Autnum + "(" + autnums[asn].Name + ")")
				delete(autnums, asn)
			} else {
				delete(autnums, asn)
				autnum, err := LoadAutnum(asn)
				if err != nil {
					raiseError("autnum/" + asn + " is invalid: " + err.Error())
				} else {
					fmt.Printf("ASN %s modified.", autnum.Autnum)
					autnums[asn] = autnum
				}
			}
		} else if strings.HasPrefix(resourcePath, "domain/") {
			domainFile := strings.TrimPrefix(resourcePath, "domain/")
			if from == nil {
				domain, err := LoadDomain(domainFile)
				if err != nil {
					raiseError("domain/" + domainFile + " is invalid: " + err.Error())
				} else {
					fmt.Println("New Domain: " + domain.Domain)
					domains[domainFile] = domain
				}
			} else if to == nil {
				fmt.Println("Deleted domain: " + domains[domainFile].Domain)
				delete(domains, domainFile)
			} else {
				delete(domains, domainFile)
				domain, err := LoadDomain(domainFile)
				if err != nil {
					raiseError("domain/" + domainFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("Domain %s modified.", domain.Domain)
					domains[domainFile] = domain
				}
			}
		} else if strings.HasPrefix(resourcePath, "inetnum/") {
			inetNumFile := strings.TrimPrefix(resourcePath, "inetnum/")
			if from == nil {
				inetNum, err := LoadInetNum(inetNumFile)
				if err != nil {
					raiseError("inetNum/" + inetNumFile + " is invalid: " + err.Error())
				} else {
					fmt.Println("New IPv4 Block: " + inetNum.CIDR)
					inetNums[inetNumFile] = inetNum
				}
			} else if to == nil {
				fmt.Println("Deleted IPv4 Block: " + inetNums[inetNumFile].CIDR)
				delete(inetNums, inetNumFile)
			} else {
				delete(inetNums, inetNumFile)
				inetNum, err := LoadInetNum(inetNumFile)
				if err != nil {
					raiseError("inetNum/" + inetNumFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv4 Block %s modified.", inetNum.CIDR)
					inetNums[inetNumFile] = inetNum
				}
			}
		} else if strings.HasPrefix(resourcePath, "inet6num/") {
			inet6NumFile := strings.TrimPrefix(resourcePath, "inet6num/")
			if from == nil {
				inet6Num, err := LoadInet6Num(inet6NumFile)
				if err != nil {
					raiseError("inet6Num/" + inet6NumFile + " is invalid: " + err.Error())
				} else {
					fmt.Println("New IPv6 Block: " + inet6Num.CIDR)
					inet6Nums[inet6NumFile] = inet6Num
				}
			} else if to == nil {
				fmt.Println("Deleted IPv6 Block: " + inet6Nums[inet6NumFile].CIDR)
				delete(inet6Nums, inet6NumFile)
			} else {
				delete(inet6Nums, inet6NumFile)
				inet6Num, err := LoadInet6Num(inet6NumFile)
				if err != nil {
					raiseError("inet6Num/" + inet6NumFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv6 Block %s modified.", inet6Num.CIDR)
					inet6Nums[inet6NumFile] = inet6Num
				}
			}
		} else if strings.HasPrefix(resourcePath, "route/") {
			routeFile := strings.TrimPrefix(resourcePath, "route/")
			if from == nil {
				route, err := LoadRoute(routeFile)
				if err != nil {
					raiseError("route/" + routeFile + " is invalid: " + err.Error())
				} else {
					fmt.Println("New IPv4 Route: " + route.CIDR)
					routes[routeFile] = route
				}
			} else if to == nil {
				fmt.Println("Deleted IPv4 Route: " + routes[routeFile].CIDR)
				delete(routes, routeFile)
			} else {
				delete(routes, routeFile)
				route, err := LoadRoute(routeFile)
				if err != nil {
					raiseError("route/" + routeFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv4 Route %s modified.", route.CIDR)
					routes[routeFile] = route
				}
			}
		} else if strings.HasPrefix(resourcePath, "route6/") {
			routeFile := strings.TrimPrefix(resourcePath, "route6/")
			if from == nil {
				route, err := LoadRoute6(routeFile)
				if err != nil {
					raiseError("route6/" + routeFile + " is invalid: " + err.Error())
				} else {
					fmt.Println("New IPv6 Route: " + route.CIDR)
					routes6[routeFile] = route
				}
			} else if to == nil {
				fmt.Println("Deleted IPv6 Route: " + routes6[routeFile].CIDR)
				delete(routes6, routeFile)
			} else {
				delete(routes6, routeFile)
				route, err := LoadRoute6(routeFile)
				if err != nil {
					raiseError("route6/" + routeFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv6 Route %s modified.", route.CIDR)
					routes6[routeFile] = route
				}
			}
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "invalid resource: "+resourcePath)
		}
	}
}
func checkout(context *Context, commit *object.Commit) {
	worktree, err := context.Repository.Worktree()
	panicErr(err)
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: commit.Hash,
	})
	panicErr(err)
}
func loadResources() {
	LoadAutnums()
	LoadDomains()
	LoadInetNums()
	LoadInet6Nums()
	LoadRoutes()
	LoadRoutes6()
}
func finalValidate() {
	autnums = map[string]*Autnum{}
	domains = map[string]*Domain{}
	inetNums = map[string]*InetNum{}
	inet6Nums = map[string]*Inet6Num{}
	routes = map[string]*Route{}
	routes6 = map[string]*Route6{}
	cidrRoot = TrieNode{}
	cidr6Root = TrieNode{}
	loadResources()
	ValidateRoutes()
	ValidateRoutes6()
}
func mergeRequestValidate() {
	context := &Context{}
	getCommit(context)
	checkout(context, context.UpstreamCommit)
	loadResources()
	checkout(context, context.PrCommit)
	checkChangedResources(context)
}
func generateROAFiles() {
	err := os.Mkdir("roa", 0750)
	panicErr(err)
	roa4 := GenerateROA()
	roa6 := GenerateROA6()
	err = os.WriteFile("roa/roa4.conf", []byte(roa4), 0666)
	panicErr(err)
	err = os.WriteFile("roa/roa6.conf", []byte(roa6), 0666)
	panicErr(err)
	err = os.WriteFile("roa/roa.conf", []byte(roa4+roa6), 0666)
	panicErr(err)
}
func main() {
	if len(os.Args) > 1 && os.Args[1] == "merge" {
		mergeRequestValidate()
	}
	if !valid {
		raiseError("registry is not valid")
		os.Exit(1)
	}
	fmt.Println("Running final check")
	finalValidate()
	if !valid {
		raiseError("registry is not valid")
		os.Exit(1)
	}
	fmt.Println("registry is valid")
	if len(os.Args) > 1 && os.Args[1] == "roa" {
		generateROAFiles()
		fmt.Println("ROA generated")
	}
}
