package elemental

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/llmos-ai/llmos/pkg/config"
)

type Elemental interface {
	Install(config.Install) error
	Upgrade(upgrade config.Upgrade) error
}

type elemental struct{}

func NewElemental() Elemental {
	return &elemental{}
}

func (r *elemental) Install(conf config.Install) error {
	var installOpts []string
	if conf.Debug {
		installOpts = append(installOpts, "--debug")
	}

	if conf.ConfigDir != "" {
		installOpts = append(installOpts, "--config-dir", conf.ConfigDir)
	}

	cmd := exec.Command("elemental", "install")
	cmd.Args = append(cmd.Args, installOpts...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	slog.Info("Running elemental install", "command", cmd.String())
	return cmd.Run()
}

func (r *elemental) Upgrade(cfg config.Upgrade) error {
	var opts = []string{"--bootloader"}

	if cfg.Debug {
		opts = append(opts, "--debug")
	}

	if cfg.UpgradeRecovery {
		opts = append(opts, "--recovery")
	}

	if cfg.Source != "" {
		opts = append(opts, "--system", cfg.Source)
	}

	cmd := exec.Command("elemental", "upgrade")
	cmd.Env = os.Environ()
	cmd.Args = append(cmd.Args, opts...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	slog.Info("Running elemental upgrade", "command", cmd.String())
	return cmd.Run()
}
