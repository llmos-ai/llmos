package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/llmos-ai/llmos/cmd/bootstrap"
	"github.com/llmos-ai/llmos/cmd/gettoken"
	"github.com/llmos-ai/llmos/cmd/info"
	"github.com/llmos-ai/llmos/cmd/install"
	"github.com/llmos-ai/llmos/cmd/probe"
	"github.com/llmos-ai/llmos/cmd/retry"
	"github.com/llmos-ai/llmos/cmd/upgrade"
	"github.com/llmos-ai/llmos/cmd/version"
)

type llmos struct {
	Debug      bool `usage:"Enable debug logging" env:"LLMOS_DEBUG"`
	DebugLevel int  `usage:"Debug log level (valid 0-9) (default 7)" env:"LLMOS_DEBUG_LEVEL"`
}

func (l *llmos) Run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

func NewRootCmd() *cobra.Command {
	l := &llmos{}
	root := cli.Command(l, cobra.Command{
		Use:   "llmos",
		Short: "LLMOS CLI Management Tool",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	})

	root.AddCommand(
		install.NewInstallCmd(root, true),
		upgrade.NewUpgradeCmd(root, true),
		bootstrap.NewBootstrap(),
		probe.NewProbe(),
		retry.NewRetry(),
		gettoken.NewGetToken(),
		info.NewInfo(),
		version.NewVersion(),
	)
	root.InitDefaultHelpCmd()
	return root
}

func (l *llmos) PersistentPre(_ *cobra.Command, _ []string) error {
	if l.Debug || l.DebugLevel > 0 {
		logging := flag.NewFlagSet("", flag.PanicOnError)
		klog.InitFlags(logging)

		level := l.DebugLevel
		if level == 0 {
			level = 6
		}
		if level > 7 {
			logrus.SetLevel(logrus.TraceLevel)
			logs.Debug = log.New(os.Stderr, "ggcr: ", log.LstdFlags)
		} else {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if err := logging.Parse([]string{
			fmt.Sprintf("-v=%d", level),
		}); err != nil {
			return err
		}
	}

	return nil
}
