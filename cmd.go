package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"golang.org/x/tools/go/loader"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: goapi [package path]
`)
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var conf loader.Config
	_, err := conf.FromArgs(flag.Args(), false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	prog, err := conf.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load packages %v\n", err)
		os.Exit(2)
	}

	var pkgNames []string
	for name, pkg := range prog.Imported {
		if pkg.Pkg.Name() == "main" {
			continue
		}
		pkgNames = append(pkgNames, name)
	}

	sort.Strings(pkgNames)

	buff := new(bytes.Buffer)
	for _, name := range pkgNames {
		formatAPI(prog.Imported[name].Pkg, &prog.Imported[name].Info, buff)
	}
	io.Copy(os.Stdout, buff)
}
