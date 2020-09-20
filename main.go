package main

import (
	"log"

	"github.com/ctison/gpm/pkg/cli"
)

var (
	version = "0.0.0"
)

func main() {
	if err := cli.New(version).Run(); err != nil {
		log.Fatal("Error: ", err)
	}
}
