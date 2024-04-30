package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llmos",
		Short: "LLMOS CLI Management Tool",
	}
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().Bool("dev", false, "Enable dev mode")
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("dev", cmd.PersistentFlags().Lookup("dev"))

	cmd.AddCommand(
		newInstallCmd(cmd, true),
		newUpgradeCmd(cmd, true),
		newVersionCmd(cmd),
	)
	cmd.SilenceUsage = true
	cmd.InitDefaultHelpCmd()
	return cmd
}
