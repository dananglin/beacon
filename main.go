package main

import (
	"fmt"
	"os"
)

var (
	binaryVersion string
	buildTime     string
	goVersion     string
	gitCommit     string
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v.\n", err)
		os.Exit(1)
	}
}

func run() error {
	return nil
}
