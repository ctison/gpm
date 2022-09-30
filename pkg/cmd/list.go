package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

type ListCommand struct {
	RootCommand *RootCommand
}

func NewCommandList(rootCommand *RootCommand) *cobra.Command {
	cmd := NewCommand()

	cmd.Aliases = []string{"l", "ls"}
	cmd.Use = "list"
	cmd.Short = "List installed assets"

	listCommand := ListCommand{
		RootCommand: rootCommand,
	}

	cmd.RunE = listCommand.RunE
	return cmd
}

func (listCommand ListCommand) RunE(cmd *cobra.Command, args []string) error {
	deps, err := listCommand.RootCommand.GPM.ListDownloadedDependencies(cmd.Context())
	if err != nil {
		return err
	}
	fmt.Println("Release assets available in cache:")
	for _, dep := range deps {
		fmt.Println(" ", dep.String())
	}
	linkedDeps, err := listCommand.RootCommand.GPM.ListLinkedDependencies(cmd.Context())
	if err != nil {
		return err
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	r := regexp.MustCompile("^" + regexp.QuoteMeta(userHomeDir+"/"))
	fmt.Println("Linked assets:")
	for _, linkedDep := range linkedDeps {
		fmt.Println(" ", r.ReplaceAllString(linkedDep.Src, "~/"), "->", r.ReplaceAllString(linkedDep.Dst, "~/"))
	}
	return nil
}
