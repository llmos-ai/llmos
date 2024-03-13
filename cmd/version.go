package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
)

func VersionCmd(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info("LLMOS cli", "version", Version, "commit", GitCommit)
		},
	}
	root.AddCommand(c)
	return c
}

var _ = VersionCmd(rootCmd)
