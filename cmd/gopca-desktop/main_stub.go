//go:build !desktop && !wails
// +build !desktop,!wails

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ERROR: This binary must be built with the 'desktop' or 'wails' build tag")
	fmt.Fprintln(os.Stderr, "Build with: go build -tags desktop")
	fmt.Fprintln(os.Stderr, "Or use: wails build")
	os.Exit(1)
}
