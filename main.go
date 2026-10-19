package main

import (
	"os"

	"github.com/leandrorosa/soundcore-q30/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.NewRoot(version).Execute(); err != nil {
		os.Exit(1)
	}
}
