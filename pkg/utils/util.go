package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/llmos-ai/llmos/pkg/config"
)

const (
	elementalConfigDir  = "/tmp/elemental"
	elementalConfigFile = "config.yaml"
)

func ValidateSource(source string) error {
	if source == "" {
		return nil
	}

	r, err := regexp.Compile(`^oci:|dir:|file:`)
	if err != nil {
		return err
	}
	if !r.MatchString(source) {
		return fmt.Errorf("source must be one of oci:|dir:|file:, current source: %s", source)
	}

	return nil
}

func ValidateRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges is required to run this command. Please run with sudo or as root user")
	}

	return nil
}

func SaveTemp(obj interface{}, prefix string) (string, error) {
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

	pterm.Info.Printf("Saved template file successfully at: %s \n%v\n", tempFile.Name(), string(bytes))

	return tempFile.Name(), nil
}

func SaveElementalConfig(elemental *config.ElementalConfig) (string, string, error) {
	err := os.MkdirAll(elementalConfigDir, os.ModePerm)
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

	pterm.Info.Printf("Saved elemental config file successfully at: %s \n%v\n", file, string(bytes))

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
