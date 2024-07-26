package upgrade

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmd2 "github.com/llmos-ai/llmos/cmd/helper"
	"github.com/llmos-ai/llmos/pkg/cli/upgrade"
	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/system"
)

func NewUpgradeCmd(root *cobra.Command, checkRoot bool) *cobra.Command {
	cfg := config.Upgrade{}
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade the LLMOS system",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd2.CheckSource(cfg.Source); err != nil {
				return err
			}
			if checkRoot {
				return cmd2.CheckRoot(viper.GetBool("dev"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := cmd2.SetupLogger(root.Context())
			cfg = setupUpgradeCfg(cfg)
			upgrade, err := upgrade.NewUpgrade(logger, cfg)
			if err != nil {
				return err
			}
			return upgrade.Run()
		},
	}
	c.Flags().StringVarP(&cfg.Source, "source", "s", "dir:/", "Set the source of the u")
	c.Flags().BoolVarP(&cfg.UpgradeRecovery, "recovery", "r", false, "Upgrade recovery system instead of the main system")
	c.Flags().StringVarP(&cfg.HostDir, "host-dir", "d", "", "Set the host directory")
	c.Flags().BoolVarP(&cfg.Force, "force", "f", false, "Force the upgrade")
	return c
}

func setupUpgradeCfg(cfg config.Upgrade) config.Upgrade {
	cfg.Debug = viper.GetBool("debug")
	cfg.Dev = viper.GetBool("dev")
	if cfg.HostDir == "" {
		cfg.HostDir = system.HostRootDir
	}
	return cfg
}
