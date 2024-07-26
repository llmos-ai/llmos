package bootstrap

import (
	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/bootstrap"
	"github.com/llmos-ai/llmos/pkg/system"
)

func NewBootstrap() *cobra.Command {
	return cli.Command(&Bootstrap{}, cobra.Command{
		Short: "Bootstrap LLMOS operator & Kubernetes",
	})
}

type Bootstrap struct {
	Force   bool   `usage:"Run bootstrap even if already bootstrapped" short:"f"`
	Config  string `usage:"Custom config path" default:"/etc/llmos/config.yaml" short:"c"`
	DataDir string `usage:"Path to llmos state dir" default:"/var/lib/llmos"`
}

func (b *Bootstrap) Run(cmd *cobra.Command, _ []string) error {
	boot := bootstrap.New(bootstrap.Config{
		Force:      b.Force,
		DataDir:    system.DataDir,
		ConfigPath: system.DefaultConfigFile,
	})
	return boot.Run(cmd.Context())
}
