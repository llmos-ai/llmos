package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/llmos-ai/llmos/cmd"
)

func main() {
	cmd := cmd.NewRootCmd()
	ctx := signals.SetupSignalHandler()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Errorf("failed to execute command: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
