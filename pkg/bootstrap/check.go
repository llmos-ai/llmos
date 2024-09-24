package bootstrap

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

func mergeConfigs(cfg Config, result config.Config) config.Config {
	logrus.Debugf("Merging config %+v to %+v", cfg, result)
	if cfg.ClusterInit {
		result.Role = config.ClusterInitRole
	}
	if cfg.Token != "" {
		result.Token = cfg.Token
	}
	if cfg.Server != "" {
		result.Server = cfg.Server
	}
	if cfg.Role != "" {
		result.Role = config.Role(cfg.Role)
	}
	if result.Role == "" && result.Server != "" && result.Token != "" {
		result.Role = config.AgentRole
	}

	if result.KubernetesVersion == "" {
		result.KubernetesVersion = cfg.KubernetesVersion
	}
	return result
}

func validateConfig(cfg *config.Config) error {
	if cfg.Role == "" && cfg.Server == "" {
		return fmt.Errorf("neither cluster-init role nor server URL is defined, skipping bootstrap")
	}

	if cfg.Role == "cluster-init" && cfg.Server != "" {
		return fmt.Errorf("cluster-init role and server URL are mutually exclusive, please select only one")
	}

	if cfg.Server != "" {
		if err := validateServerURL(cfg.Server); err != nil {
			return fmt.Errorf("invalid server URL: %v", err)
		}
	}

	if cfg.Server != "" && cfg.Token == "" {
		return fmt.Errorf("server URL is defined but token is not, skipping bootstrap")
	}

	return nil
}

func validateServerURL(serverURL string) error {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %v", err)
	}

	// Check scheme
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid server URL: scheme must be https")
	}

	// Check port
	port := parsedURL.Port()
	if port != "6443" && port != "9345" {
		return fmt.Errorf("invalid server URL: port must be 6443 or 9345")
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Check if the server url is ready
	url := fmt.Sprintf("%s/ping", strings.TrimSuffix(serverURL, "/"))
	resp, err := retryClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check server URL: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to validate server url: %v", err)
	}

	if string(body) != "pong" || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server url is not ready: %s", string(body))
	}

	return nil
}
