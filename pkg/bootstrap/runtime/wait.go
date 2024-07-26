package runtime

import (
	"fmt"
	"os"

	"github.com/llmos-ai/llmos/utils/cmd"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
)

func ToWaitNodeReadyInstruction(nodeName, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location of %s: %w", os.Args[0], err)
	}

	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name: "wait-node-ready",
			Args: []string{"retry", kubectl.Command(k8sVersion), "wait",
				"--for=condition=Ready", fmt.Sprintf("node/%s", nodeName)},
			Env:     kubectl.Env(k8sVersion),
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}

func ToWaitSystemAgentActiveInstruction(k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location of %s: %w", os.Args[0], err)
	}

	target := fmt.Sprintf("%s-agent.service", config.GetRuntime(k8sVersion))

	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "wait-agent-node-ready",
			Args:    []string{"retry", "systemctl", "is-active", target},
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}
