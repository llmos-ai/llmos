package info

import (
	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/bootstrap"
	"github.com/llmos-ai/llmos/pkg/system"
)

func NewInfo() *cobra.Command {
	return cli.Command(&Info{}, cobra.Command{
		Short: "Print installation versions",
	})
}

type Info struct {
}

func (b *Info) Run(cmd *cobra.Command, _ []string) error {
	r := bootstrap.New(bootstrap.Config{
		DataDir:    system.DataDir,
		ConfigPath: system.DefaultConfigFile,
	})
	return r.Info(cmd.Context())
}
