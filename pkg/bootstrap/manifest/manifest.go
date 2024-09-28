package manifest

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	cmd2 "github.com/llmos-ai/llmos/utils/cmd"
	"github.com/llmos-ai/llmos/utils/randomtoken"
	"github.com/llmos-ai/llmos/utils/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	llmosCfg "github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/images"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
)

const (
	localK8sStateTypeName = "k8s.llmos.ai/cluster-state"
)

func GetNodeName(config *llmosCfg.Config) (string, error) {
	nodeName := config.NodeName
	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("looking up hostname: %w", err)
		}
		nodeName = strings.Split(hostname, ".")[0]
	}
	return strings.ToLower(nodeName), nil
}

func ToBootstrapFile(config *llmosCfg.Config, path string, runtime llmosCfg.Runtime) (*applyinator.File, error) {
	nodeName, err := GetNodeName(config)
	if err != nil {
		return nil, err
	}

	token := config.Token
	if token == "" {
		token, err = randomtoken.Generate()
		if err != nil {
			return nil, err
		}
	}

	resources := config.Resources
	return ToFile(append(resources,
		llmosCfg.GenericMap{
			Data: map[string]interface{}{
				"kind":       "Node",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": nodeName,
					"labels": map[string]interface{}{
						"llmos.ai/managed": "true",
					},
				},
			},
		}, llmosCfg.GenericMap{
			Data: map[string]interface{}{
				"kind":       "Namespace",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "llmos-system",
					"annotations:": map[string]interface{}{
						"llmos.ai/bootstrap-version": config.LLMOSOperatorVersion,
					},
				},
			},
		}, llmosCfg.GenericMap{
			Data: map[string]interface{}{
				"kind":       "Secret",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "local-k8s-state",
					"namespace": "llmos-system",
					"labels": map[string]interface{}{
						"llmos.ai/k8s-provider": runtime,
					},
				},
				"type": localK8sStateTypeName,
				"data": map[string]interface{}{
					"serverToken": []byte(token),
					"agentToken":  []byte(token),
				},
			},
		}), path)
}
func ToFile(resources []llmosCfg.GenericMap, path string) (*applyinator.File, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	var objs []runtime.Object
	for _, resource := range resources {
		objs = append(objs, &unstructured.Unstructured{
			Object: resource.Data,
		})
	}

	data, err := yaml.ToBytes(objs)
	if err != nil {
		return nil, err
	}

	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    path,
	}, nil
}

func ToInstruction(imageOverride, systemDefaultRegistry, k8sVersion, dataDir string) (*applyinator.OneTimeInstruction, error) {
	bootstrap := GetBootstrapManifests(dataDir)
	cmd, err := cmd2.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "bootstrap",
			Image:   images.GetRuntimeInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
			Args:    []string{"retry", kubectl.Command(k8sVersion), "apply", "--validate=false", "-f", bootstrap},
			Command: cmd,
			Env:     kubectl.Env(k8sVersion),
		},
		SaveOutput: true,
	}, nil
}

func GetBootstrapManifests(dataDir string) string {
	return fmt.Sprintf("%s/bootstrapmanifests/llmos.yaml", dataDir)
}
