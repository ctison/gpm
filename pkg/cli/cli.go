package cli

import (
	"os"

	"github.com/ctison/gpm/pkg/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CLI holds the context for the command line.
type CLI struct {
	log    *log.Logger
	ng     *engine.Engine
	root   *cobra.Command
	search search
}

// Create a CLI to run.
func New(version string) *CLI {
	cli := &CLI{}

	// Setup logger.
	cli.log = &log.Logger{
		Out:   os.Stderr,
		Level: log.DebugLevel,
		Formatter: &log.TextFormatter{
			DisableTimestamp: true,
		},
	}

	// Setup engine.
	ng, err := engine.New(cli.log)
	if err != nil {
		cli.log.Fatal("Error: ", err)
	}
	cli.ng = ng

	// Setup root command.
	cli.root = newCommand()
	cli.root.Use = "gpm"
	cli.root.Version = version
	cli.root.PersistentFlags().Bool("help", false, "Print help and exit")
	cli.root.PersistentPostRunE = func(_ *cobra.Command, _ []string) error {
		return cli.ng.Stop()
	}

	// Setup subcommands.
	cli.setupList()
	cli.setupInstall()
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
