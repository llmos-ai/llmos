package images

import (
	"fmt"
	"strings"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

const (
	defaultRuntimeImagePrefix = "rancher/system-agent-installer"
	defaultSystemImagePrefix  = "llmos-ai/system-installer"
)

func GetLLMOSInstallerImage(imageOverride, imagePrefix, operatorVersion string) string {
	return getInstallerImage(imageOverride, imagePrefix, "llmos-operator", operatorVersion)
}

func GetRuntimeInstallerImage(imageOverride, imagePrefix, kubernetesVersion string) string {
	if imagePrefix == "" {
		imagePrefix = defaultRuntimeImagePrefix
	}
	return getInstallerImage(imageOverride, imagePrefix, string(config.GetRuntime(kubernetesVersion)), kubernetesVersion)
}

func getInstallerImage(imageOverride, imagePrefix, component, version string) string {
	if imageOverride != "" {
		return imageOverride
	}

	if imagePrefix == "" {
		imagePrefix = defaultSystemImagePrefix
	}

	tag := strings.ReplaceAll(version, "+", "-")
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s-%s:%s", imagePrefix, component, tag)
}
