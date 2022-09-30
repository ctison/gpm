package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ctison/gpm/pkg/gpm"
	"github.com/ctison/gpm/pkg/tui"
	"github.com/spf13/cobra"
)

type InstallCommand struct {
	RootCommand *RootCommand
	Name        string
}

func NewCommandInstall(rootCommand *RootCommand) *cobra.Command {
	cmd := NewCommand()

	cmd.Aliases = []string{"i"}
	cmd.Use = "install [OWNER/]REPOSITORY[@TAG][:ARTIFACT[,...]] [...]"
	cmd.Short = "Install release assets"

	installCommand := InstallCommand{
		RootCommand: rootCommand,
	}

	cmd.LocalFlags().StringVarP(&installCommand.Name, "name", "n", "", "Override the name of the installed executable (Default to repository name)")

	cmd.RunE = installCommand.RunE
	return cmd
}

func (installCommand InstallCommand) RunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("install from manifest is not yet implemented")
	}

	deps, err := gpm.ConvertDependenciesStrings(args...)

	if err != nil {
		return fmt.Errorf("failed to parse the argument(s): %w", err)
	}

	if debug := installCommand.RootCommand.Debug; debug != "" {
		f, err := tea.LogToFile(debug, "")
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", debug, err)
		}
		defer f.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	m, err := tea.NewProgram(tui.NewInstallModel(*installCommand.RootCommand.GPM, deps...)).StartReturningModel()
	if err != nil {
		return err
	}

	if m.(tui.InstallModel).Errored() {
		os.Exit(1)
	}

	return nil

}
