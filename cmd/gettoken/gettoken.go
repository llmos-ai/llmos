package gettoken

import (
	"fmt"

	"github.com/llmos-ai/llmos/utils/cli"
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/token"
)

func NewGetToken() *cobra.Command {
	return cli.Command(&GetToken{}, cobra.Command{
		Short: "Print token to join nodes to the cluster",
	})
}

type GetToken struct{}

func (p *GetToken) Run(cmd *cobra.Command, _ []string) error {
	str, err := token.GetLocalToken(cmd.Context())
	if err != nil {
		return err
	}
	fmt.Print(str)
	return nil
}
