package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ctison/gpm/pkg/gpm"
	"github.com/ctison/gpm/pkg/tui"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableAutoGenTag:     true,
		DisableFlagsInUseLine: true,
	}
	return cmd
}

type RootCommand struct {
	Verbose   bool
	Debug     string
	Config    string
	HomePath  string
	StorePath string
	BinPath   string
	GPM       *gpm.GPM
}

func NewRootCommand() *cobra.Command {
	cobraCommand := NewCommand()
	cobraCommand.Short = "Github Package Manager is a tool to install assets from Github releases"

	rootCommand := &RootCommand{}

	cobraCommand.AddCommand(NewCommandInstall(rootCommand), NewCommandList(rootCommand))

	cobraCommand.PersistentFlags().BoolVarP(&rootCommand.Verbose, "verbose", "v", false, "Enable verbosity")
	cobraCommand.PersistentFlags().StringVarP(&rootCommand.Debug, "debug", "d", "", "File path to write debugging logs")
	cobraCommand.PersistentFlags().StringVarP(&rootCommand.Config, "config", "c", "gpm.yaml", "Configuration file")
	cobraCommand.PersistentFlags().StringVar(&rootCommand.HomePath, "home-dir", "", "Base path used to compute store dir and bin dir (Defaults to ~)")
	cobraCommand.PersistentFlags().StringVar(&rootCommand.StorePath, "store-dir", "", "Base path used to store downloaded assets (Defaults to ~/.local/share/gpm)")
	cobraCommand.PersistentFlags().StringVar(&rootCommand.BinPath, "bin-dir", "", "Directory where symlinks to executables will be created (Defaults to ~/.local/bin)")

	cobraCommand.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		rootCommand.GPM = gpm.NewGPM(
			gpm.WithHomePath(rootCommand.HomePath),
			gpm.WithBinPath(rootCommand.BinPath),
			gpm.WithStorePath(rootCommand.StorePath),
		)
		return nil
	}

	cobraCommand.RunE = rootCommand.RunE

	return cobraCommand
}

func (rc RootCommand) RunE(_ *cobra.Command, _ []string) error {
	return tea.NewProgram(tui.NewDashboardModel()).Start()
}
