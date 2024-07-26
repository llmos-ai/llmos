package probe

import (
	"fmt"
	"time"

	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/probe"
)

func NewProbe() *cobra.Command {
	return cli.Command(&Probe{}, cobra.Command{
		Short:  "Run plan probes",
		Hidden: true,
	})
}

type Probe struct {
	Interval string `usage:"Polling interval to run probes" default:"2s" short:"i"`
	File     string `usage:"Plan file" default:"/var/lib/llmos/plan/plan.json" short:"f"`
}

func (p *Probe) Run(cmd *cobra.Command, _ []string) error {
	interval, err := time.ParseDuration(p.Interval)
	if err != nil {
		return fmt.Errorf("parsing duration %s: %w", p.Interval, err)
	}

	return probe.RunProbes(cmd.Context(), p.File, interval)
}
