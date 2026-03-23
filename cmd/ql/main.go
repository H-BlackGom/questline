package main

import (
	"os"

	"github.com/H-BlackGom/questline/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
