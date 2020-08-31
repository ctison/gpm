package cmd

import (
	"github.com/ctison/gpm/pkg/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommandInstall() *cobra.Command {
	cmd := NewCommand()
	cmd.Use = "install [USER/]REPOSITORY[@VERSION] [ARTIFACT [...]]"
	cmd.Short = "Install one or more release artifacts from a Github repository"
	cmd.Aliases = []string{"i"}
	cmd.RunE = install
	cmd.Args = cobra.MinimumNArgs(1)
	return cmd
}

func install(cmd *cobra.Command, args []string) error {
	if err := viper.GetViper().ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	release, err := engine.NewRelease(args[0])
	if err != nil {
		return err
	}
	ng, err := engine.NewEngine()
	if err != nil {
		return err
	}
	_, err = ng.Install(*release)
	if err != nil {
		return err
	}
	return nil
}
