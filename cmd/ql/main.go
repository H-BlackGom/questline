package main

import (
	"errors"
	"os"

	"github.com/H-BlackGom/questline/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		switch {
		case errors.Is(err, cli.ErrInvalidInput):
			os.Exit(2)
		case errors.Is(err, cli.ErrQuestNotFound):
			os.Exit(3)
		case errors.Is(err, cli.ErrDatabase):
			os.Exit(4)
		default:
			os.Exit(1)
		}
	}
}
