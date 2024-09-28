package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/llmos-ai/llmos/utils/data"
	"github.com/llmos-ai/llmos/utils/data/convert"
	"github.com/llmos-ai/llmos/utils/yaml"
	"github.com/rancher/wharfie/pkg/registries"
	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/applyinator/image"
)

var (
	implicitPaths = []string{
		"/usr/share/oem/llmos/bootstrap/config.yaml",
		"/usr/share/llmos/bootstrap/config.yaml",
		// llmos userdata
		"/oem/userdata",
		// llmos installation yip config
		"/oem/99_custom.yaml",
		// llmos oem location
		"/oem/llmos/config.yaml",
		// Standard cloud-config
		"/var/lib/cloud/instance/user-data.txt",
	}

	manifests = []string{
		"/usr/share/oem/llmos/manifests",
		"/usr/share/llmos/manifests",
		"/etc/llmos/manifests",
		// llmos OEM
		"/oem/llmos/manifests",
	}
)

type GenericMap struct {
	Data map[string]interface{} `json:"-"`
}

type Config struct {
	RuntimeConfig
	KubernetesVersion    string `json:"kubernetesVersion,omitempty"`
	LLMOSOperatorVersion string `json:"llmosOperatorVersion,omitempty"`
	ChartRepo            string `json:"chartRepo,omitempty"`

	LLMOSOperatorValues map[string]interface{}           `json:"llmosOperatorValues,omitempty"`
	PreInstructions     []applyinator.OneTimeInstruction `json:"preInstructions,omitempty"`
	PostInstructions    []applyinator.OneTimeInstruction `json:"postInstructions,omitempty"`
	Resources           []GenericMap                     `json:"manifest,omitempty"`

	RuntimeInstallerImage string               `json:"runtimeInstallerImage,omitempty"`
	LLMOSInstallerImage   string               `json:"llmosInstallerImage,omitempty"`
	GlobalImageRegistry   string               `json:"globalImageRegistry,omitempty"`
	Registries            *registries.Registry `json:"registries,omitempty"`
	ImageUtility          *image.Utility       `json:"imageUtility,omitempty"`
}

func paths() (result []string) {
	for _, file := range implicitPaths {
		result = append(result, file)

		files, err := os.ReadDir(file)
		if err != nil {
			continue
		}

		for _, entry := range files {
			if isYAML(entry.Name()) {
				result = append(result, filepath.Join(file, entry.Name()))
			}
		}
	}
	return
}

func Load(path string) (result Config, err error) {
	var (
		values = map[string]interface{}{}
	)

	if err = populatedSystemResources(&result); err != nil {
		return result, err
	}

	for _, file := range paths() {
		newValues, err := mergeFile(values, file)
		if err == nil {
			values = newValues
		} else {
			logrus.Infof("failed to parse %s, skipping file: %v", file, err)
		}
	}

	if path != "" {
		values, err = mergeFile(values, path)
		if err != nil {
			return
		}
	}

	err = convert.ToObj(values, &result)
	if err != nil {
		return
	}

	return result, err
}

func populatedSystemResources(config *Config) error {
	resources, err := loadResources(manifests...)
	if err != nil {
		return err
	}
	config.Resources = append(config.Resources, resources...)

	return nil
}

func isYAML(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}

func loadResources(dirs ...string) (result []GenericMap, _ error) {
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !isYAML(path) {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			objs, err := yaml.ToObjects(f)
			if err != nil {
				return err
			}

			for _, obj := range objs {
				apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
				if apiVersion == "" || kind == "" {
					continue
				}
				data, err := convert.EncodeToMap(obj)
				if err != nil {
					return err
				}
				result = append(result, GenericMap{
					Data: data,
				})
			}

			return nil
		})
		if os.IsNotExist(err) {
			continue
		}
	}

	return
}

func mergeFile(result map[string]interface{}, file string) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	files, err := dotDFiles(file)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{}
	if len(bytes) > 0 {
		logrus.Infof("Loading config file [%s]", file)
		if err := yaml.Unmarshal(bytes, &values); err != nil {
			return nil, err
		}
	}

	if v, ok := values["llmos"].(map[string]interface{}); ok {
		values = v
	}

	result = data.MergeMapsConcatSlice(result, values)
	for _, file := range files {
		result, err = mergeFile(result, file)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func dotDFiles(basefile string) (result []string, _ error) {
	files, err := os.ReadDir(basefile + ".d")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}
		result = append(result, filepath.Join(basefile+".d", file.Name()))
	}
	return
}
