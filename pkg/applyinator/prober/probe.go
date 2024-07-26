package prober

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	k8sprobe "k8s.io/kubernetes/pkg/probe"
	k8shttp "k8s.io/kubernetes/pkg/probe/http"
)

type HTTPGetAction struct {
	URL        string `json:"url,omitempty"`
	Insecure   bool   `json:"insecure,omitempty"`
	ClientCert string `json:"clientCert,omitempty"`
	ClientKey  string `json:"clientKey,omitempty"`
	CACert     string `json:"caCert,omitempty"`
}

type Probe struct {
	Name                string        `json:"name,omitempty"`
	InitialDelaySeconds int           `json:"initialDelaySeconds,omitempty"` // default 0
	TimeoutSeconds      int           `json:"timeoutSeconds,omitempty"`      // default 1
	SuccessThreshold    int           `json:"successThreshold,omitempty"`    // default 1
	FailureThreshold    int           `json:"failureThreshold,omitempty"`    // default 3
	HTTPGetAction       HTTPGetAction `json:"httpGet,omitempty"`
}

type ProbeStatus struct {
	Healthy      bool `json:"healthy,omitempty"`
	SuccessCount int  `json:"successCount,omitempty"`
	FailureCount int  `json:"failureCount,omitempty"`
}

func DoProbe(probe Probe, probeStatus *ProbeStatus, initial bool) error {
	logrus.Tracef("Running probe %+v", probe)
	if initial {
		initialDelayDuration := time.Duration(probe.InitialDelaySeconds) * time.Second
		logrus.Debugf("[Probe: %s] Sleeping for %.0f seconds before running probe", probe.Name, initialDelayDuration.Seconds())
		time.Sleep(initialDelayDuration)
	}

	var k8sProber k8shttp.Prober

	if probe.HTTPGetAction.Insecure {
		k8sProber = k8shttp.New(false)
	} else {
		tlsConfig := tls.Config{}
		if probe.HTTPGetAction.ClientCert != "" && probe.HTTPGetAction.ClientKey != "" {
			clientCert, err := tls.LoadX509KeyPair(probe.HTTPGetAction.ClientCert, probe.HTTPGetAction.ClientKey)
			if err != nil {
				logrus.Errorf("error loading x509 client cert/key for probe %s (%s/%s): %v", probe.Name, probe.HTTPGetAction.ClientCert, probe.HTTPGetAction.ClientKey, err)
			}
			tlsConfig.Certificates = []tls.Certificate{clientCert}
		}

		caCertPool, err := GetSystemCertPool(probe.Name)
		if err != nil || caCertPool == nil {
			caCertPool = x509.NewCertPool()
			logrus.Errorf("error loading system cert pool for probe (%s): %v", probe.Name, err)
		}

		if probe.HTTPGetAction.CACert != "" {
			logrus.Debugf("[DoProbe] adding CA certificate [%s] for probe (%s)", probe.HTTPGetAction.CACert, probe.Name)
			caCert, err := os.ReadFile(probe.HTTPGetAction.CACert)
			if err != nil {
				logrus.Errorf("error loading CA cert for probe (%s) %s: %v", probe.Name, probe.HTTPGetAction.CACert, err)
			}
			if !caCertPool.AppendCertsFromPEM(caCert) {
				logrus.Errorf("error while appending ca cert to pool for probe %s", probe.Name)
			}
		}

		tlsConfig.RootCAs = caCertPool
		k8sProber = k8shttp.NewWithTLSConfig(&tlsConfig, false)
	}

	probeURL, err := url.Parse(probe.HTTPGetAction.URL)
	if err != nil {
		return err
	}

	probeRequest, err := k8shttp.NewProbeRequest(probeURL, http.Header{})
	if err != nil {
		return err
	}

	probeDuration := time.Duration(probe.TimeoutSeconds) * time.Second
	logrus.Tracef("[Probe: %s] timeout duration: %.0f seconds", probe.Name, probeDuration.Seconds())

	probeResult, output, err := k8sProber.Probe(probeRequest, probeDuration)
	if err != nil {
		logrus.Errorf("error while running probe (%s): %v", probe.Name, err)
		return err
	}

	logrus.Debugf("[Probe: %s] output was %s", probe.Name, output)

	var successThreshold, failureThreshold int

	if probe.SuccessThreshold == 0 {
		logrus.Tracef("[Probe: %s] Setting success threshold to default", probe.Name)
		successThreshold = 1
	} else {
		logrus.Tracef("[Probe: %s] Setting success threshold to %d", probe.Name, probe.SuccessThreshold)
		successThreshold = probe.SuccessThreshold
	}

	if probe.FailureThreshold == 0 {
		logrus.Tracef("[Probe: %s] Setting failure threshold to default", probe.Name)
		failureThreshold = 3
	} else {
		logrus.Tracef("Setting failure threshold to %d", probe.FailureThreshold)
		failureThreshold = probe.FailureThreshold
	}

	switch probeResult {
	case k8sprobe.Success:
		logrus.Debugf("[Probe: %s] succeeded", probe.Name)
		if probeStatus.SuccessCount < successThreshold {
			probeStatus.SuccessCount = probeStatus.SuccessCount + 1
			if probeStatus.SuccessCount >= successThreshold {
				probeStatus.Healthy = true
			}
		}
		probeStatus.FailureCount = 0
	default:
		logrus.Debugf("[Probe: %s] failed", probe.Name)
		if probeStatus.FailureCount < failureThreshold {
			probeStatus.FailureCount = probeStatus.FailureCount + 1
			if probeStatus.FailureCount >= failureThreshold {
				probeStatus.Healthy = false
			}
		}
		probeStatus.SuccessCount = 0
	}

	return nil
}

// GetSystemCertPool returns a x509.CertPool that contains the
// root CA certificates if they are present at runtime
func GetSystemCertPool(probeName string) (*x509.CertPool, error) {
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
		logrus.Errorf("[GetSystemCertPoolUnix] error loading system cert pool for probe (%s): %v", probeName, err)
	}
	if caCertPool == nil {
		return nil, fmt.Errorf("[GetSystemCertPoolWindows] x509 returned a nil certpool for probe (%s)", probeName)
	}
	return caCertPool, nil
}

func DoProbes(probes map[string]Probe, probeStatuses map[string]ProbeStatus, initial bool) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for probeName, probe := range probes {
		wg.Add(1)
		go func(probeName string, probe Probe) {
			defer wg.Done()
			logrus.Debugf("[Prober] (%s) running probe", probeName)

			mu.Lock()
			defer mu.Unlock()

			//defer logrus.Tracef("[Prober] (%s) retrieving existing probe status from map if existing", probeName)
			probeStatus, ok := probeStatuses[probeName]
			//mu.Unlock()
			if !ok {
				logrus.Tracef("[Prober] (%s) probe status was not present in map, initializing", probeName)
				probeStatus = ProbeStatus{}
			}

			probe.Name = probeName
			if err := DoProbe(probe, &probeStatus, initial); err != nil {
				logrus.Errorf("error running probe %s", probeName)
			}

			//mu.Lock()
			logrus.Tracef("[Prober] (%s) writing probe status to map", probeName)
			probeStatuses[probeName] = probeStatus
			//mu.Unlock()
		}(probeName, probe)
	}

	// wait for all probes to complete
	wg.Wait()
}
