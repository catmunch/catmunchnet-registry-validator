package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
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
				}
			} else if to == nil {
				index := -1
				for i, autnum := range autnums {
					if autnum.Autnum == asn {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find ASN %s", asn))
				}
				fmt.Println("Deleted ASN: " + autnums[index].Autnum + "(" + autnums[index].Name + ")")
				autnums = append(autnums[:index], autnums[index+1:]...)
			} else {
				index := -1
				for i, autnum := range autnums {
					if autnum.Autnum == asn {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find ASN %s", asn))
				}
				autnums = append(autnums[:index], autnums[index+1:]...)
				autnum, err := LoadAutnum(asn)
				if err != nil {
					raiseError("autnum/" + asn + " is invalid: " + err.Error())
				} else {
					fmt.Printf("ASN %s modified.", autnum.Autnum)
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
				}
			} else if to == nil {
				index := -1
				for i, domain := range domains {
					if domain.Domain == domainFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find domain %s", domainFile))
				}
				fmt.Println("Deleted domain: " + domains[index].Domain)
				domains = append(domains[:index], domains[index+1:]...)
			} else {
				index := -1
				for i, domain := range domains {
					if domain.Domain == domainFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find domain %s", domainFile))
				}
				domains = append(domains[:index], domains[index+1:]...)
				domain, err := LoadDomain(domainFile)
				if err != nil {
					raiseError("domain/" + domainFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("Domain %s modified.", domain.Domain)
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
				}
			} else if to == nil {
				index := -1
				for i, inetNum := range inetNums {
					if strings.ReplaceAll(inetNum.CIDR, "/", "_") == inetNumFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find inetnum %s", inetNumFile))
				}
				fmt.Println("Deleted IPv4 Block: " + inetNums[index].CIDR)
				inetNums = append(inetNums[:index], inetNums[index+1:]...)
			} else {
				index := -1
				for i, inetNum := range inetNums {
					if strings.ReplaceAll(inetNum.CIDR, "/", "_") == inetNumFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find inetnum %s", inetNumFile))
				}
				inetNums = append(inetNums[:index], inetNums[index+1:]...)
				inetNum, err := LoadInetNum(inetNumFile)
				if err != nil {
					raiseError("inetNum/" + inetNumFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv4 Block %s modified.", inetNum.CIDR)
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
				}
			} else if to == nil {
				index := -1
				for i, inet6Num := range inet6Nums {
					if strings.ReplaceAll(inet6Num.CIDR, "/", "_") == inet6NumFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find inet6Num %s", inet6NumFile))
				}
				fmt.Println("Deleted IPv6 Block: " + inet6Nums[index].CIDR)
				inet6Nums = append(inet6Nums[:index], inet6Nums[index+1:]...)
			} else {
				index := -1
				for i, inet6Num := range inet6Nums {
					if strings.ReplaceAll(inet6Num.CIDR, "/", "_") == inet6NumFile {
						index = i
						break
					}
				}
				if index == -1 {
					panicErr(fmt.Errorf("cannot find inet6Num %s", inet6NumFile))
				}
				inet6Nums = append(inet6Nums[:index], inet6Nums[index+1:]...)
				inet6Num, err := LoadInet6Num(inet6NumFile)
				if err != nil {
					raiseError("inet6Num/" + inet6NumFile + " is invalid: " + err.Error())
				} else {
					fmt.Printf("IPv6 Block %s modified.", inet6Num.CIDR)
				}
			}
			// TODO: } else if strings.HasPrefix(resourcePath, "route/") {
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
}
func mergeRequestValidate() {
	context := &Context{}
	getCommit(context)
	checkout(context, context.UpstreamCommit)
	loadResources()
	checkout(context, context.PrCommit)
	checkChangedResources(context)
}
func main() {
	if len(os.Args) > 1 && os.Args[1] == "merge" {
		mergeRequestValidate()
	} else {
		loadResources()
	}
	if !valid {
		raiseError("registry is not valid")
		os.Exit(1)
	}
	fmt.Println("registry is valid")
}
