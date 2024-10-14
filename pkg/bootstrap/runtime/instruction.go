package runtime

import (
	"fmt"
	"os"

	"github.com/llmos-ai/llmos/utils/cmd"
	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/images"
	"github.com/llmos-ai/llmos/pkg/utils"
)

const (
	llmosKubeconfigPath = "/etc/llmos/kubeconfig.yaml"
)

func ToInstruction(cfg *config.Config, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	runtime := config.GetRuntime(k8sVersion)
	env := addRuntimeEnvConfig(runtime, cfg, k8sVersion)
	logrus.Debugf("runtime %s instruction envs: %+v", runtime, env)

	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:  fmt.Sprintf("install-%s", runtime),
			Env:   env,
			Image: images.GetRuntimeInstallerImage(cfg.RuntimeInstallerImage, cfg.GlobalImageRegistry, k8sVersion),
		},
		SaveOutput: true,
	}, nil
}

func addRuntimeEnvConfig(runtime config.Runtime, cfg *config.Config, k8sVersion string) []string {
	var env []string
	env = utils.AddEnv(env, "RESTART_STAMP",
		images.GetRuntimeInstallerImage(cfg.RuntimeInstallerImage, cfg.GlobalImageRegistry, k8sVersion))
	// define join role to either server to agent
	// k3s:  https://github.com/k3s-io/k3s/blob/38e8b01b8f9bb6709df90ac5839e4579115664a7/install.sh#L172-L183
	// rke2: https://github.com/rancher/rke2/blob/96041884eaf06bcd1a4586b429b01ba51561e651/install.sh#L25-L27
	if cfg.Role == config.AgentRole && runtime == config.RuntimeK3S {
		env = utils.AddEnv(env, "K3S_URL", cfg.Server)
		env = utils.AddEnv(env, "K3S_TOKEN", cfg.Token)
	} else if cfg.Role == config.AgentRole && runtime == config.RuntimeRKE2 {
		env = utils.AddEnv(env, "INSTALL_RKE2_TYPE", "agent")
	}

	if runtime == config.RuntimeRKE2 {
		env = utils.AddEnv(env, "RKE2_ENABLE_SERVICELB", "true")
	}

	return env
}

func CopyKubeConfigInstruction(k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	runtime := config.GetRuntime(k8sVersion)
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    fmt.Sprintf("symlink-kubeconfig-%s", runtime),
			Args:    []string{"retry", "ln", "-sf", GetKubeconfigPath(runtime), llmosKubeconfigPath},
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}

func GetKubeconfigPath(runtime config.Runtime) string {
	return fmt.Sprintf("/etc/rancher/%s/%s.yaml", runtime, runtime)
}
