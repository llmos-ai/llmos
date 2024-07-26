package role

import "strings"

func IsEtcd(role string) bool {
	return strings.Contains(role, "server") ||
		strings.Contains(role, "etcd") ||
		strings.Contains(role, "cluster-init")
}

func IsControlPlane(role string) bool {
	return strings.Contains(role, "server") ||
		strings.Contains(role, "cluster-init") ||
		strings.Contains(role, "control-plane") ||
		strings.Contains(role, "controlplane")
}

func IsWorker(role string) bool {
	return strings.Contains(role, "worker") ||
		strings.Contains(role, "agent") ||
		strings.Contains(role, "cluster-init") ||
		strings.Contains(role, "server")
}
