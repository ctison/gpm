package main

import (
	"os"

	"github.com/ctison/gpm/pkg/cmd"
)

var (
	version = "0.0.0"
)

func main() {
	cmd := cmd.NewRootCommand()
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
