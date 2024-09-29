package info

import (
	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/bootstrap"
)

func NewInfo() *cobra.Command {
	return cli.Command(&Info{}, cobra.Command{
		Short: "Print installation versions",
	})
}

type Info struct {
	Config  string `usage:"Custom config file path" default:"/etc/llmos/config.yaml" short:"c" env:"LLMOS_CONFIG_FILE"`
	DataDir string `usage:"Path to llmos state dir" default:"/var/lib/llmos" env:"LLMOS_DATA_DIR"`
}

func (b *Info) Run(cmd *cobra.Command, _ []string) error {
	r := bootstrap.New(bootstrap.Config{
		DataDir:    b.DataDir,
		ConfigPath: b.Config,
	})
	return r.Info(cmd.Context())
}
