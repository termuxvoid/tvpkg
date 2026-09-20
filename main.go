package main

import (
	"fmt"
	"os"

	"tvpkg/cmd"
)

// ldflags injection (see Makefile / CI):
//
//	-X tvpkg/cmd.version=...
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
