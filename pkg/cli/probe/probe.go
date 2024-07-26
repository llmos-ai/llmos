package probe

import (
	"fmt"
	"os"
	"strings"

	"github.com/llmos-ai/llmos/utils/cmd"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/applyinator/prober"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/role"
)

var probes = map[string]prober.Probe{
	"kube-apiserver": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL:        "https://127.0.0.1:6443/readyz",
			CACert:     "/var/lib/rancher/%s/server/tls/server-ca.crt",
			ClientCert: "/var/lib/rancher/%s/server/tls/client-kube-apiserver.crt",
			ClientKey:  "/var/lib/rancher/%s/server/tls/client-kube-apiserver.key",
		},
	},
	"kube-scheduler": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL:      "https://127.0.0.1:10259/healthz",
			Insecure: true,
		},
	},
	"kube-controller-manager": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL:      "https://127.0.0.1:10257/healthz",
			Insecure: true,
		},
	},
	"kubelet": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL: "http://127.0.0.1:10248/healthz",
		},
	},
}

func replaceRuntime(str string, runtime config.Runtime) string {
	if !strings.Contains(str, "%s") {
		return str
	}
	return fmt.Sprintf(str, runtime)
}

func ProbesForJoin(cfg *config.RuntimeConfig) map[string]prober.Probe {
	if role.IsControlPlane(string(cfg.Role)) {
		return AllProbes(config.RuntimeUnknown)
	}
	return replaceRuntimeForProbes(map[string]prober.Probe{
		"kubelet": probes["kubelet"],
	}, config.RuntimeUnknown)
}

func AllProbes(runtime config.Runtime) map[string]prober.Probe {
	return replaceRuntimeForProbes(probes, runtime)
}

func replaceRuntimeForProbes(probes map[string]prober.Probe, runtime config.Runtime) map[string]prober.Probe {
	result := map[string]prober.Probe{}
	for k, v := range probes {
		// we don't know the runtime to find the file
		if runtime == config.RuntimeUnknown && (v.HTTPGetAction.CACert+
			v.HTTPGetAction.ClientCert+
			v.HTTPGetAction.ClientKey) != "" {
			continue
		}
		v.HTTPGetAction.CACert = replaceRuntime(v.HTTPGetAction.CACert, runtime)
		v.HTTPGetAction.ClientCert = replaceRuntime(v.HTTPGetAction.ClientCert, runtime)
		v.HTTPGetAction.ClientKey = replaceRuntime(v.HTTPGetAction.ClientKey, runtime)
		result[k] = v
	}
	return result
}

func ToInstruction() (*applyinator.OneTimeInstruction, error) {
	cmd, err := cmd.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "probes",
			Args:    []string{"probe"},
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}
