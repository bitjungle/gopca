package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("CompLab PCA Toolkit v0.1.0")
	if len(os.Args) < 2 {
		fmt.Println("Usage: complab-cli <command>")
		os.Exit(1)
	}
}
