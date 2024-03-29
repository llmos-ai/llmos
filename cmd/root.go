package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llmos",
		Short: "LLMOS is a lightweight Large Language Model(LLM) based Operating System",
	}
	cmd.PersistentFlags().String("config-dir", "", "Set config directory")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().Bool("quiet", false, "Disable output")
	cmd.PersistentFlags().String("logfile", "", "Config logfile")
	_ = viper.BindPFlag("config-dir", cmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("logfile", cmd.PersistentFlags().Lookup("logfile"))

	cmd.AddCommand(
		newInstallCmd(cmd),
		newServeCmd(cmd),
		newVersionCmd(cmd),
	)
	cmd.SilenceUsage = true
	return cmd
}
