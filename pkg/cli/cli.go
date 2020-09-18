package cli

import "github.com/spf13/cobra"

// CLI holds the context for the command line.
type CLI struct {
	root   *cobra.Command
	search search
}

// Create a CLI to run.
func New(version string) *CLI {
	cli := &CLI{}
	cli.root = newCommand()
	cli.root.Use = "gpm"
	cli.root.Version = version
	cli.root.PersistentFlags().Bool("help", false, "Print help and exit")
	cli.setupSearch()
	return cli
}

// Run CLI.
func (cli *CLI) Run() error {
	return cli.root.Execute()
}

// Create an empty command prefilled with defaults.
func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableAutoGenTag:     true,
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
		SilenceUsage:          true,
	}
	return cmd
}
