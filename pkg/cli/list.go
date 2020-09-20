package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *CLI) setupList() {
	cmd := newCommand()
	cmd.Use = "list"
	cmd.Aliases = []string{"l", "ls"}
	cmd.Run = cli.List
	cli.root.AddCommand(cmd)
}

func (cli *CLI) List(cmd *cobra.Command, args []string) {
	assets, err := cli.ng.List()
	if err != nil {
		cli.log.Fatal("Error: ", err)
	}
	for _, asset := range assets {
		fmt.Printf("%s -> %s\n", asset.LinkName, asset.Path)
	}
}
