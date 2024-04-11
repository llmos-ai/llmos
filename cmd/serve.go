package cmd

import (
	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/serve"
)

var server serve.ServeOptions

func newServeCmd(root *cobra.Command) *cobra.Command {
	server = serve.NewServe()
	c := &cobra.Command{
		Use:   "serve",
		Short: "Run the LLM model and UI server",
		Args:  cobra.MaximumNArgs(1),
		RunE:  serveAction,
	}
	c.Flags().StringVarP(&server.Model, "model", "m", "mistral:7b", "Default model to serve")
	c.Flags().StringVarP(&server.Namespace, "namespace", "n", "llmos", "Namespace for containerd")
	return c
}

func serveAction(cmd *cobra.Command, args []string) error {
	return server.StartServe(cmd.Context())
}
