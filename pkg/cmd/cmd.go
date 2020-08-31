package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableAutoGenTag:     true,
		DisableFlagsInUseLine: true,
	}
	return cmd
}

func NewRootCommand() *cobra.Command {
	viper.SetConfigName("config.yaml")
	viper.AddConfigPath("/etc/gpm/")
	viper.AddConfigPath("$HOME/.gpm/")
	cmd := NewCommand()
	cmd.Short = "Github Package Manager is a tool to install artifacts from Github releases"
	cmd.AddCommand(NewCommandSearch(), NewCommandInstall())
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbosity")
	cmd.Flags().Bool("debug", false, "Enable debug. Implies verbosity.")
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}
	return cmd
}
