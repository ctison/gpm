package cmd

import (
	"fmt"

	"github.com/ctison/gpm/pkg/engine"
	"github.com/spf13/cobra"
)

func NewCommandSearch() *cobra.Command {
	cmd := NewCommand()
	cmd.Use = "search [[USER/]REPOSITORY[@VERSION] [...]]"
	cmd.Short = "Search"
	cmd.Aliases = []string{"s"}
	cmd.RunE = search
	return cmd
}

func search(cmd *cobra.Command, args []string) error {
	engine, err := engine.NewEngine()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		locals, err := engine.List("")
		if err != nil {
			return err
		}
		for _, local := range locals {
			remote, err := engine.SearchRemote(local)
			if err != nil {
				return err
			}
			if local.Version != remote.Version {
				fmt.Printf("local:%s -> remote:%s", local, remote)
			}
		}
	}
	return nil
}
