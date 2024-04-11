package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/install"
	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/utils"
)

func newInstallCmd(root *cobra.Command) *cobra.Command {
	installCfg := &InstallConfig{}
	c := &cobra.Command{
		Use:   "install",
		Short: "Run the LLMOS installation",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateSource(installCfg.Source); err != nil {
				return err
			}
			return utils.ValidateRoot()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.LLMOSConfig{}
			if err := install.AskInstall(cfg); err != nil {
				pterm.Error.Printf("Failed to install, error: %s\n", err)
				return err
			}
			return nil
		},
	}
	c.Flags().StringVarP(&installCfg.Source, "source", "s", "", "Source of the LLMOS installation")
	c.Flags().BoolP("silent", "y", false, "Run the installation in silent mode")
	c.Flags().BoolP("reboot", "r", true, "Reboot the system after installation")
	return c
}

type InstallConfig struct {
	Source string `json:"source"`
}
