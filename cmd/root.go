package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llmos",
		Short: "LLMOS is a lightweight LLM Linux distribution",
	}
	cmd.PersistentFlags().String("config-dir", "", "Set config directory")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().Bool("quiet", false, "Disable output")
	cmd.PersistentFlags().String("logfile", "", "Config logfile")
	_ = viper.BindPFlag("config-dir", cmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("logfile", cmd.PersistentFlags().Lookup("logfile"))
	return cmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Errorf("Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = NewRootCmd()
