package main

import (
	"log/slog"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/llmos-ai/llmos/cmd"
)

func main() {
	cmd := cmd.NewRootCmd()
	ctx := signals.SetupSignalHandler()
	if err := cmd.ExecuteContext(ctx); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
