package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/version"
)

func newVersionCmd(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info(fmt.Sprintf("LLMOS CLI version: %s", version.GetFriendlyVersion()))
		},
	}
	return c
}
