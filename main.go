package main

import (
	"log/slog"
	"os"

	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/oneblock-ai/llmos/cmd"
)

func main() {
	ctx := controllerruntime.SetupSignalHandler()
	cmd := cmd.NewRootCmd()
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
