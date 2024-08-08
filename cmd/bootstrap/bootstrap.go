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
	Force       bool   `usage:"Run bootstrap even if already bootstrapped" short:"f" env:"LLMOS_BOOTSTRAP_FORCE"`
	Config      string `usage:"Custom config path" default:"/etc/llmos/config.yaml" short:"c" env:"LLMOS_CONFIG"`
	DataDir     string `usage:"Path to llmos state dir" default:"/var/lib/llmos" env:"LLMOS_DATA_DIR"`
	ClusterInit bool   `usage:"Bootstrap cluster-init role" env:"LLMOS_CLUSTER_INIT"`
	Server      string `usage:"Server to connect to" env:"LLMOS_SERVER"`
	Token       string `usage:"Token to use for authentication" env:"LLMOS_TOKEN"`
}

func (b *Bootstrap) Run(cmd *cobra.Command, _ []string) error {
	boot := bootstrap.New(bootstrap.Config{
		Force:       b.Force,
		DataDir:     system.DataDir,
		ConfigPath:  system.DefaultConfigFile,
		Server:      b.Server,
		Token:       b.Token,
		ClusterInit: b.ClusterInit,
	})
	return boot.Run(cmd.Context())
}
