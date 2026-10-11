//go:build ignore
// +build ignore

// create_mod builds a Go module zip (the same layout proxy.golang.org serves)
// from a source directory.
//
// Usage: create_mod <module-path> <version> <source-dir> <output-zip>
package main

import (
	"log"
	"os"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatalf("usage: %s <module-path> <version> <source-dir> <output-zip>", os.Args[0])
	}
	modPath, version, dir, out := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	if err := zip.CreateFromDir(f, module.Version{Path: modPath, Version: version}, dir); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
