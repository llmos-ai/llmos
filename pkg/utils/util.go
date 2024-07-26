package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils/log"
)

const (
	elementalConfigDir  = "/tmp/elemental"
	elementalConfigFile = "config.yaml"
)

func SaveTemp(obj interface{}, prefix string, logger log.Logger, print bool) (string, error) {
	tempFile, err := os.CreateTemp("/tmp", fmt.Sprintf("%s.", prefix))
	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	if _, err = tempFile.Write(bytes); err != nil {
		return "", err
	}
	if err = tempFile.Close(); err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Saved %s file successfully", prefix), "fileName", tempFile.Name())
	if logger.IsDebug() || print {
		pterm.Info.Print(string(bytes))
	}

	return tempFile.Name(), nil
}

func SaveElementalConfig(elemental *elemental.ElementalConfig, logger log.Logger) (string, string, error) {
	if _, err := os.Stat(elementalConfigDir); os.IsNotExist(err) {
		err := os.MkdirAll(elementalConfigDir, os.ModePerm)
		if err != nil {
			return "", "", err
		}
	}

	_, err := os.Create(filepath.Join(elementalConfigDir, elementalConfigFile))
	if err != nil {
		return "", "", err
	}

	bytes, err := yaml.Marshal(elemental)
	if err != nil {
		return "", "", err
	}

	file := filepath.Join(elementalConfigDir, elementalConfigFile)
	err = os.WriteFile(file, bytes, os.ModePerm)
	if err != nil {
		return "", "", err
	}

	logger.Info("Saved elemental config file successfully", "fileName", file)
	if logger.IsDebug() {
		pterm.Info.Print(string(bytes))
	}

	return elementalConfigDir, file, nil
}

func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

func SetEnv(env []string) {
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) >= 2 {
			os.Setenv(pair[0], pair[1])
		}
	}
}

func IsRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func IsK8sPod() bool {
	if env := os.Getenv("KUBERNETES_SERVICE_HOST"); env != "" {
		return true
	}
	return false
}
func AddEnv(env []string, key, value string) []string {
	return append(env, fmt.Sprintf("%s=%s", key, value))
}
