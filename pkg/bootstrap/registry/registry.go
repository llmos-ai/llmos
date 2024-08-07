package registry

import (
	"encoding/base64"
	"fmt"

	"github.com/rancher/wharfie/pkg/registries"
	"sigs.k8s.io/yaml"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

func ToFile(registry *registries.Registry, runtime config.Runtime) (*applyinator.File, error) {
	if registry == nil || len(registry.Configs) == 0 {
		return nil, nil
	}

	data, err := yaml.Marshal(registry)
	if err != nil {
		return nil, err
	}

	return &applyinator.File{
		Content:     base64.StdEncoding.EncodeToString(data),
		Path:        GetRuntimeConfigFile(runtime),
		Permissions: "0400",
	}, nil

}

func GetRuntimeConfigFile(runtime config.Runtime) string {
	return fmt.Sprintf("/etc/rancher/%s/registries.yaml", runtime)
}
