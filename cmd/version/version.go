package version

import (
	"fmt"

	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/version"
)

type Version struct {
}

func NewVersion() *cobra.Command {
	return cli.Command(&Version{}, cobra.Command{
		Use:                "version",
		Short:              "Print the version",
		DisableFlagParsing: true,
	})
}

func (v *Version) Run(_ *cobra.Command, _ []string) error {
	fmt.Println(version.GetFriendlyVersion())
	return nil
}
