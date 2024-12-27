package images

import (
	"fmt"
	"strings"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

const (
	defaultGhcrRegistry         = "ghcr.io"
	defaultRuntimeImagePrefix   = "rancher/system-agent-installer"
	defaultInstallerImagePrefix = "llmos-ai/system-installer"

	DefaultVolcMirrorRegistry = "llmos-ai-cn-beijing.cr.volces.com"
	AliSystemDefaultRegistry  = "registry.cn-hangzhou.aliyuncs.com"
)

func GetLLMOSInstallerImage(imageOverride, registry, mirror, operatorVersion string) string {
	if registry == "" && mirror != "" {
		registry = DefaultVolcMirrorRegistry
	}
	return getInstallerImage(imageOverride, registry, defaultInstallerImagePrefix, "llmos-operator", operatorVersion)
}

func GetRuntimeInstallerImage(imageOverride, registry, mirror, kubernetesVersion string) string {
	if registry == "" && mirror != "" {
		registry = AliSystemDefaultRegistry
	} else {
		registry = "docker.io"
	}

	return getInstallerImage(imageOverride, registry, defaultRuntimeImagePrefix,
		string(config.GetRuntime(kubernetesVersion)), kubernetesVersion)
}

func getInstallerImage(imageOverride, registry, imagePrefix, component, version string) string {
	if imageOverride != "" {
		return imageOverride
	}

	if registry == "" {
		registry = defaultGhcrRegistry
	}

	if imagePrefix == "" {
		imagePrefix = defaultInstallerImagePrefix
	}

	tag := strings.ReplaceAll(version, "+", "-")
	if tag == "" {
		tag = "latest"
	}

	return fmt.Sprintf("%s/%s-%s:%s", registry, imagePrefix, component, tag)
}
