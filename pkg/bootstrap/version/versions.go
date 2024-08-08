package version

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	cachedOperatorVersion = map[string]string{}
	cachedK8sVersion      = map[string]string{}
	cachedLock            sync.Mutex
	redirectClient        = &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

func getVersionOrURL(urlFormat, def, version string) (_ string, isURL bool) {
	if version == "" {
		version = def
	}

	if strings.HasPrefix(version, "v") && len(strings.Split(version, ".")) > 2 {
		return version, false
	}

	channelURL := version
	if !strings.HasPrefix(channelURL, "https://") &&
		!strings.HasPrefix(channelURL, "http://") {
		if strings.HasSuffix(channelURL, "-head") || strings.Contains(channelURL, "/") {
			return channelURL, false
		}
		channelURL = fmt.Sprintf(urlFormat, version)
	}

	return channelURL, true
}

func K8sVersion(kubernetesVersion string) (string, error) {
	cachedLock.Lock()
	defer cachedLock.Unlock()

	cached, ok := cachedK8sVersion[kubernetesVersion]
	if ok {
		return cached, nil
	}

	urlFormat := "https://update.k3s.io/v1-release/channels/%s"
	if strings.HasSuffix(kubernetesVersion, ":k3s") {
		kubernetesVersion = strings.TrimSuffix(kubernetesVersion, ":k3s")
	} else if strings.HasSuffix(kubernetesVersion, ":rke2") {
		urlFormat = "https://update.rke2.io/v1-release/channels/%s"
		kubernetesVersion = strings.TrimSuffix(kubernetesVersion, ":rke2")
	}

	versionOrURL, isURL := getVersionOrURL(urlFormat, "stable", kubernetesVersion)
	if !isURL {
		return versionOrURL, nil
	}

	resp, err := redirectClient.Get(versionOrURL)
	if err != nil {
		return "", fmt.Errorf("getting channel version from (%s): %w", versionOrURL, err)
	}
	defer resp.Body.Close()

	url, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("getting channel version URL from (%s): %w", versionOrURL, err)
	}

	resolved := path.Base(url.Path)
	cachedK8sVersion[kubernetesVersion] = resolved
	logrus.Infof("Resolving Kubernetes version [%s] to %s from %s ", kubernetesVersion, resolved, versionOrURL)
	return resolved, nil
}

func OperatorVersion(repo, version string) (string, error) {
	cachedLock.Lock()
	defer cachedLock.Unlock()

	cached, ok := cachedOperatorVersion[version]
	if ok {
		return cached, nil
	}

	if repo == "" {
		repo = "latest"
	}

	versionOrURL, isURL := getVersionOrURL("https://releases.1block.ai/charts/%s/index.yaml", repo, version)
	if !isURL {
		return versionOrURL, nil
	}

	resp, err := http.Get(versionOrURL)
	if err != nil {
		return "", fmt.Errorf("getting llmos-operator channel version from (%s): %w", versionOrURL, err)
	}
	defer resp.Body.Close()

	index := &chartIndex{}
	if err := yaml.NewDecoder(resp.Body).Decode(index); err != nil {
		return "", fmt.Errorf("unmarshalling llmos-operator channel version from (%s): %w", versionOrURL, err)
	}

	versions := index.Entries["llmos-operator"]
	if len(versions) == 0 {
		return "", fmt.Errorf("failed to find version for llmos-operator chart at (%s)", versionOrURL)
	}

	ver := versions[0].AppVersion

	logrus.Infof("Resolving llmos-operator version [%s] to %s from %s ", version, ver, versionOrURL)
	cachedOperatorVersion[version] = ver
	return ver, nil
}

type chartIndex struct {
	Entries map[string][]struct {
		Version    string `yaml:"version"`
		AppVersion string `yaml:"appVersion"`
	} `yaml:"entries"`
}

func GetClusterK8sAndOperatorVersions(serverURL, token string) (string, string, error) {
	if serverURL == "" || token == "" {
		return "", "", fmt.Errorf("server and token must be provided")
	}

	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid server URL: %v", err)
	}

	url := fmt.Sprintf("https://%s:%s/v1-cluster/cluster-info", parsedURL.Hostname(), "30443")

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	standardClient := retryClient.StandardClient() // *http.Client

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := standardClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to get cluster info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cluster info body: %v", err)
	}

	clusterInfo := &clusterInfo{}
	err = json.Unmarshal(body, clusterInfo)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	k8sVersion := clusterInfo.K8sVersion
	operatorVersion := clusterInfo.LLMOSOperatorVersion

	cachedK8sVersion[k8sVersion] = k8sVersion
	cachedOperatorVersion[operatorVersion] = operatorVersion

	return k8sVersion, operatorVersion, nil
}

type clusterInfo struct {
	K8sVersion           string `json:"k8sVersion"`
	LLMOSOperatorVersion string `json:"llmosOperatorVersion"`
}
