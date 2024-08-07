package plan

import (
	"os"

	"github.com/llmos-ai/llmos/utils/data/convert"
	"github.com/llmos-ai/llmos/utils/randomtoken"
	"gopkg.in/yaml.v3"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/runtime"
	"github.com/llmos-ai/llmos/pkg/bootstrap/version"
)

func assignTokenIfUnset(cfg *config.Config) error {
	if cfg.Token != "" {
		return nil
	}

	token, err := existingToken(cfg)
	if err != nil {
		return err
	}

	if token == "" {
		token, err = randomtoken.Generate()
		if err != nil {
			return err
		}
	}

	cfg.Token = token
	return nil
}

func existingToken(cfg *config.Config) (string, error) {
	k8sVersion, err := version.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return "", err
	}

	cfgFile := runtime.GetConfigLocation(config.GetRuntime(k8sVersion))
	data, err := os.ReadFile(cfgFile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	configMap := map[string]interface{}{}
	if err = yaml.Unmarshal(data, &configMap); err != nil {
		return "", err
	}

	return convert.ToString(configMap["token"]), nil
}
