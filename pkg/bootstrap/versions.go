package bootstrap

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"runtime"
	"strings"

	data2 "github.com/llmos-ai/llmos/utils/data"
	"github.com/llmos-ai/llmos/utils/data/convert"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
)

const prettyPrefix = "PRETTY_NAME="

func (l *LLMOS) getExistingVersions(ctx context.Context) (operatorVersion, k8sVersion, llmOSVersion string) {
	osVersion := getOSVersion()
	kubeConfig, err := kubectl.GetKubeconfig("")
	if err != nil {
		logrus.Debugf("failed to get kubeconfig file %s", err.Error())
		return "", "", osVersion
	}

	data, err := os.ReadFile(kubeConfig)
	if err != nil {
		logrus.Debugf("failed to read kubeconfig file %s", err.Error())
		return "", "", osVersion
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		logrus.Debugf("failed to get rest config %s", err.Error())
		return "", "", osVersion
	}

	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logrus.Debugf("failed to get k8s client %s", err.Error())
		return "", "", osVersion
	}

	return getOperatorVersion(ctx, k8s), getK8sVersion(ctx, k8s), osVersion
}

func getOperatorVersion(ctx context.Context, k8s kubernetes.Interface) string {
	secrets, err := k8s.CoreV1().Secrets("llmos-system").List(ctx, metav1.ListOptions{
		LabelSelector: "name=llmos-operator,status=deployed",
	})
	if err != nil || len(secrets.Items) == 0 {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(string(secrets.Items[0].Data["release"]))
	if err != nil {
		return ""
	}

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	release := map[string]interface{}{}
	if err := json.NewDecoder(gz).Decode(&release); err != nil {
		return ""
	}

	version := convert.ToString(data2.GetValueN(release, "chart", "metadata", "version"))
	if version == "" {
		return ""
	}

	return "v" + version
}

func getK8sVersion(ctx context.Context, k8s kubernetes.Interface) string {
	nodes, err := k8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/control-plane=true",
	})
	if err != nil || len(nodes.Items) == 0 {
		return ""
	}
	return nodes.Items[0].Status.NodeInfo.KubeletVersion
}

func getOSVersion() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		logrus.Errorf("failed to read os-relase file, error: %v\n", err)
		return ""
	}

	scan := bufio.NewScanner(bytes.NewBuffer(data))
	for scan.Scan() {
		if strings.HasPrefix(scan.Text(), prettyPrefix) {
			os := strings.TrimSuffix(strings.TrimPrefix(scan.Text(), prettyPrefix), "-"+runtime.GOARCH)
			if len(os) > 0 && os[0] == '"' {
				os = os[1:]
			}
			if len(os) > 0 && os[len(os)-1] == '"' {
				os = os[:len(os)-1]
			}
			return os
		}
	}
	return ""
}
