package cli

import (
	"github.com/ctison/gpm/pkg/engine"
	"github.com/spf13/cobra"
)

func (cli *CLI) setupInstall() {
	cmd := newCommand()
	cmd.Use = "install " + engine.Format + " ..."
	cmd.Run = cli.Install
	cmd.Aliases = []string{"i"}
	cmd.Args = cobra.MinimumNArgs(1)
	cli.root.AddCommand(cmd)
}

func (cli *CLI) Install(cmd *cobra.Command, args []string) {
	// Setup engine.
	for _, arg := range args {
		asset, err := engine.ParseAsset(arg)
		if err != nil {
			cli.log.Fatal("Error: ", err)
		}
		if err := cli.ng.GuessOwner(asset); err != nil {
			cli.log.Fatal("Error: ", err)
		}
		if err := cli.ng.GuessVersion(asset); err != nil {
			cli.log.Fatal("Error: ", err)
		}
		if err := cli.ng.GuessName(asset); err != nil {
			cli.log.Fatal("Error: ", err)
		}
		if err := cli.ng.Install(asset); err != nil {
			cli.log.Fatal("Error: ", err)
		}
		cli.log.Infof("Asset: %s", asset)
	}
}
