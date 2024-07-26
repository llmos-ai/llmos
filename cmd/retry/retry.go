package retry

import (
	"time"

	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/retry"
)

func NewRetry() *cobra.Command {
	return cli.Command(&Retry{}, cobra.Command{
		Short:              "Retry command until it succeeds",
		DisableFlagParsing: true,
		Hidden:             true,
	})
}

type Retry struct {
	SleepFirst bool `usage:"Sleep 5 seconds before running command"`
}

func (p *Retry) Run(cmd *cobra.Command, args []string) error {
	if p.SleepFirst {
		time.Sleep(5 * time.Second)
	}
	return retry.Retry(cmd.Context(), 15*time.Second, args)
}
