package main

import "github.com/dnephin/gorepack/repack"

func main() {
	err := repack.Rename(".", repack.RenameOpts{
		Package: "github.com/docker/docker",
		Exclude:  []string{"vendor", ".git"},
	})
	if err != nil {
		panic(err)
	}
}
