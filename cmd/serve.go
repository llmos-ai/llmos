package cmd

import (
	"github.com/spf13/cobra"
)

func ServeCmd(root *cobra.Command) *cobra.Command {
	cfg := &ServeConfig{}
	c := &cobra.Command{
		Use:   "serve",
		Short: "Run the LLM model and UI server",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	root.AddCommand(c)
	c.Flags().StringVarP(&cfg.Model, "model", "m", "mistral:7b", "Default model to serve")
	return c
}

type ServeConfig struct {
	Model string `json:"model"`
}

var _ = ServeCmd(rootCmd)
