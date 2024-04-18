package config

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"gopkg.in/yaml.v3"
)

func ReadLLMOSConfigFile(file string) (*Config, error) {
	var err error
	var data []byte

	// check if is valid file
	_, err = os.Stat(file)
	if err == nil {
		data, err = GetLocalLLMOSConfig(file)
		if err != nil {
			return nil, err
		}
	}

	// check if source is a valid url
	if strings.Contains(file, "http") {
		url, err := url.Parse(file)
		if err != nil {
			slog.Debug("invalid source url", "file", file)
			return nil, err
		}
		data, err = GetRemoteLLMOSConfig(url.String())
		if err != nil {
			return nil, fmt.Errorf("error reading LLMOS config file from url: %s", err.Error())
		}

	}

	if len(data) > 0 {
		return LoadLLMOSConfig(data)
	}

	return nil, fmt.Errorf("invalid LLMOS config file: %s", file)
}

func GetLocalLLMOSConfig(path string) ([]byte, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading local LLMOS config file: %s", err.Error())
	}
	return bytes, nil
}

func GetRemoteLLMOSConfig(url string) ([]byte, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	resp, err := retryClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the HTTP response is a success (2xx) or success-like code (3xx)
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, fmt.Errorf("url response status code is invalid: %d", resp.StatusCode)
}

func LoadLLMOSConfig(yamlBytes []byte) (*Config, error) {
	cfg := NewLLMOSConfig()
	if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	return cfg, nil
}
