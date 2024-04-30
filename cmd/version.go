package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/version"
)

func newVersionCmd(_ *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetFriendlyVersion())
		},
	}
	return c
}
