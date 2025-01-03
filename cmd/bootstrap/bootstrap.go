package bootstrap

import (
	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/bootstrap"
)

func NewBootstrap() *cobra.Command {
	return cli.Command(&Bootstrap{}, cobra.Command{
		Short: "Bootstrap LLMOS operator & Kubernetes",
	})
}

// Bootstrap defines the command to bootstrap LLMOS
//
//nolint:all
type Bootstrap struct {
	Force             bool   `usage:"Run bootstrap even if already bootstrapped" short:"f" env:"LLMOS_BOOTSTRAP_FORCE"`
	Config            string `usage:"Custom config file path" default:"/etc/llmos/config.yaml" short:"c" env:"LLMOS_CONFIG_FILE"`
	DataDir           string `usage:"Path to llmos state dir" default:"/var/lib/llmos" env:"LLMOS_DATA_DIR"`
	Server            string `usage:"Server url to connect to" env:"LLMOS_SERVER"`
	Role              string `usage:"The node role to join the cluster" enum:"server,agent" short:"r" env:"LLMOS_ROLE"`
	Token             string `usage:"Token to use for join the cluster" env:"LLMOS_TOKEN"`
	ClusterInit       bool   `usage:"Bootstrap cluster-init role" env:"LLMOS_CLUSTER_INIT"`
	KubernetesVersion string `usage:"Default kubernetes version to bootstrap" env:"LLMOS_KUBERNETES_VERSION" default:"v1.31.3+k3s1"`
	Mirror            string `usage:"Specify the mirror registry for installation" enum:"cn" env:"LLMOS_MIRROR"`
}

func (b *Bootstrap) Run(cmd *cobra.Command, _ []string) error {
	boot := bootstrap.New(bootstrap.Config{
		Force:             b.Force,
		DataDir:           b.DataDir,
		ConfigPath:        b.Config,
		Server:            b.Server,
		Role:              b.Role,
		Token:             b.Token,
		ClusterInit:       b.ClusterInit,
		KubernetesVersion: b.KubernetesVersion,
		Mirror:            b.Mirror,
	})
	return boot.Run(cmd.Context())
}
