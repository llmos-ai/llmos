package operator

import (
	"fmt"
	"os"

	"github.com/llmos-ai/llmos/utils/cmd"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
)

func ToWaitOperatorInstruction(k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}

	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name: "wait-llmos-operator",
			Args: []string{"retry", kubectl.Command(k8sVersion), "-n", "llmos-system",
				"rollout", "status", "-w", "deploy/llmos-operator"},
			Env:     kubectl.Env(k8sVersion),
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}

func ToWaitOperatorWebhookInstruction(k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name: "wait-operator-webhook",
			Args: []string{"retry", kubectl.Command(k8sVersion), "-n", "llmos-system",
				"rollout", "status", "-w", "deploy/llmos-operator-webhook"},
			Env:     kubectl.Env(k8sVersion),
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}

func ToWaitSUCInstruction(_, _, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "wait-system-upgrade-controller",
			Args:    []string{"retry", kubectl.Command(k8sVersion), "-n", "system-upgrade", "rollout", "status", "-w", "deploy/system-upgrade-controller"},
			Env:     kubectl.Env(k8sVersion),
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}
