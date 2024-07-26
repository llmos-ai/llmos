package runtime

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/llmos-ai/llmos/utils/data/convert"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

var (
	normalizeNames = map[string]string{
		"tlsSans":         "tls-san",
		"nodeName":        "node-name",
		"internalAddress": "internal-address",
		"taints":          "node-taint",
		"labels":          "node-label",
	}
)

func ToTokenFile(token, dataDir string) (*applyinator.File, error) {
	tokenByte := []byte(fmt.Sprintf("%s\n", token))
	return &applyinator.File{
		Content:     base64.StdEncoding.EncodeToString(tokenByte),
		Path:        fmt.Sprintf("%s/token", dataDir),
		Permissions: "600",
	}, nil
}

func ToBootstrapFile(config *config.RuntimeConfig, runtime config.Runtime, server string) (*applyinator.File, error) {
	data, err := ToConfig(config, server)
	if err != nil {
		return nil, err
	}
	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    GetConfigLocation(runtime),
	}, nil
}

func ToConfig(config *config.RuntimeConfig, server string) ([]byte, error) {
	configObjects := []interface{}{
		config.ConfigValues,
	}

	configObjects = append(configObjects, config)
	result := map[string]interface{}{}
	for _, data := range configObjects {
		mapData, err := convert.EncodeToMap(data)
		if err != nil {
			return nil, err
		}

		delete(mapData, "extraConfig")
		delete(mapData, "role")
		for oldKey, newKey := range normalizeNames {
			value, ok := mapData[oldKey]
			if !ok {
				continue
			}
			delete(mapData, oldKey)
			mapData[newKey] = value
		}
		for k, v := range mapData {
			newKey := strings.ReplaceAll(convert.ToYAMLKey(k), "_", "-")
			result[newKey] = v
		}

		if len(server) == 0 {
			result["cluster-init"] = "true"
		}
	}

	logrus.Debugf("generated LLMOS config: %+v\n", result)
	return yaml.Marshal(result)
}

func GetConfigLocation(runtime config.Runtime) string {
	return fmt.Sprintf("/etc/rancher/%s/config.yaml.d/40-llmos.yaml", runtime)
}
