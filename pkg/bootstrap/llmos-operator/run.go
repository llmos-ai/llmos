package operator

import (
	"encoding/base64"
	"fmt"

	"github.com/llmos-ai/llmos/utils/data"
	"sigs.k8s.io/yaml"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/images"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
)

var defaultValues = map[string]interface{}{
	"operator": map[string]interface{}{
		"apiserver": map[string]interface{}{
			"replicaCount": 1,
		},
	},
}

func GetOperatorValues(dataDir string) string {
	return fmt.Sprintf("%s/llmos-operator/values.yaml", dataDir)
}

func ToFile(cfg *config.Config, dataDir string) (*applyinator.File, error) {
	values := data.MergeMaps(defaultValues, map[string]interface{}{
		"global": map[string]interface{}{
			"imageRegistry": cfg.GlobalImageRegistry,
		},
	})
	values = data.MergeMaps(values, cfg.LLMOSOperatorValues)

	data, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshalling LLMOS-Operator values.yaml: %w", err)
	}

	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    GetOperatorValues(dataDir),
	}, nil
}

func ToInstruction(imageOverride, systemDefaultRegistry, k8sVersion, operatorVersion, dataDir string) (*applyinator.OneTimeInstruction, error) {
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:  "llmos-operator",
			Image: images.GetLLMOSInstallerImage(imageOverride, systemDefaultRegistry, operatorVersion),
			Env:   append(kubectl.Env(k8sVersion), fmt.Sprintf("LLMOS_VALUES=%s", GetOperatorValues(dataDir))),
		},
		SaveOutput: true,
	}, nil
}
