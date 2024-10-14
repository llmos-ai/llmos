package config

import (
	"strings"
)

var (
	RuntimeRKE2    Runtime = "rke2"
	RuntimeK3S     Runtime = "k3s"
	RuntimeUnknown Runtime = "unknown"

	ClusterInitRole Role = "cluster-init"
	ServerRole      Role = "server"
	AgentRole       Role = "agent"
)

type Runtime string
type Role string

type RuntimeConfig struct {
	Role            Role                   `json:"role,omitempty"`
	Server          string                 `json:"server,omitempty"`
	SANS            []string               `json:"tlsSans,omitempty"`
	NodeName        string                 `json:"nodeName,omitempty"`
	Address         string                 `json:"address,omitempty"`
	InternalAddress string                 `json:"internalAddress,omitempty"`
	Taints          []string               `json:"taints,omitempty"`
	Labels          []string               `json:"labels,omitempty"`
	Token           string                 `json:"token,omitempty"`
	ConfigValues    map[string]interface{} `json:"extraConfig,omitempty"`
}

func GetRuntime(kubernetesVersion string) Runtime {
	if isRKE2(kubernetesVersion) {
		return RuntimeRKE2
	}

	if isK3s(kubernetesVersion) {
		return RuntimeK3S
	}
	return RuntimeUnknown
}

func isRKE2(kubernetesVersion string) bool {
	return strings.Contains(kubernetesVersion, string(RuntimeRKE2))
}

func isK3s(kubernetesVersion string) bool {
	return strings.Contains(kubernetesVersion, string(RuntimeK3S))
}
