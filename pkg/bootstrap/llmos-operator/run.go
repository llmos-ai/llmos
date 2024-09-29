package operator

import (
	"encoding/base64"
	"fmt"
	"os"

	helmv1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	cmd2 "github.com/llmos-ai/llmos/utils/cmd"
	"github.com/llmos-ai/llmos/utils/data"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/images"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
	"github.com/llmos-ai/llmos/pkg/constants"
)

const (
	helmAPIVersion     = "helm.cattle.io/v1"
	helmConfigKindName = "HelmChartConfig"
)

var defaultValues = map[string]interface{}{
	"operator": map[string]interface{}{
		"apiserver": map[string]interface{}{
			"service": map[string]interface{}{
				"type":          "LoadBalancer",
				"httpsPort":     8443,
				"httpsNodePort": 30443,
			},
		},
	},
}

func GetOperatorChartConfigPath(dataDir string) string {
	return fmt.Sprintf("%s/charts/llmos-operator-config.yaml", dataDir)
}

func ToFile(cfg *config.Config, dataDir string) (*applyinator.File, error) {
	values := data.MergeMaps(defaultValues, map[string]interface{}{
		"global": map[string]interface{}{
			"imageRegistry": cfg.GlobalImageRegistry,
		},
	})
	values = data.MergeMaps(values, cfg.LLMOSOperatorValues)

	valuesData, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshalling llmos-operator values.yaml: %w", err)
	}

	config := helmv1.HelmChartConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmAPIVersion,
			Kind:       helmConfigKindName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.LLMOSOperatorName,
			Namespace: constants.SystemNamespace,
		},
		Spec: helmv1.HelmChartConfigSpec{
			ValuesContent: string(valuesData),
		},
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshalling llmos-operator HelmChartConfig: %w", err)
	}

	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    GetOperatorChartConfigPath(dataDir),
	}, nil
}

func ToInstruction(imageOverride, systemDefaultRegistry, k8sVersion,
	operatorVersion string) (*applyinator.OneTimeInstruction, error) {
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:  "install-llmos-operator",
			Image: images.GetLLMOSInstallerImage(imageOverride, systemDefaultRegistry, operatorVersion),
			Env:   kubectl.Env(k8sVersion),
		},
		SaveOutput: true,
	}, nil
}

func ToChartConfigInstruction(k8sVersion, dataDir string) (*applyinator.OneTimeInstruction, error) {
	file := GetOperatorChartConfigPath(dataDir)
	cmd, err := cmd2.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "apply-operator-chart-config",
			Args:    []string{"retry", kubectl.Command(k8sVersion), "apply", "-f", file},
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}
