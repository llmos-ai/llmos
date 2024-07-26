package main

import (
	"os"

	"github.com/pterm/pterm"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/llmos-ai/llmos/cmd"
)

func main() {
	cmd := cmd.NewRootCmd()
	ctx := signals.SetupSignalHandler()
	cmd.SilenceErrors = true
	if err := cmd.ExecuteContext(ctx); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
