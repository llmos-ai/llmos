package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/applyinator/prober"
)

func RunProbes(_ context.Context, planFile string, interval time.Duration) error {
	f, err := os.Open(planFile)
	if err != nil {
		return fmt.Errorf("opening plan %s: %w", planFile, err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logrus.Fatalln(err)
		}
	}()

	plan := &applyinator.Plan{}
	if err = json.NewDecoder(f).Decode(plan); err != nil {
		return err
	}

	if len(plan.Probes) == 0 {
		logrus.Infof("No probes defined in %s", planFile)
		return nil
	}
	logrus.Infof("Running probes defined in %s", planFile)

	probeStatuses := make(map[string]prober.ProbeStatus)
	initial := true

	for {
		newProbeStatuses := map[string]prober.ProbeStatus{}
		for k, v := range probeStatuses {
			newProbeStatuses[k] = v
		}

		allGood := true
		prober.DoProbes(plan.Probes, newProbeStatuses, initial)

		for probeName, probeStatus := range probeStatuses {
			if !probeStatus.Healthy {
				allGood = false
			}

			oldProbeStatus, ok := probeStatuses[probeName]
			if !ok || oldProbeStatus.Healthy != probeStatus.Healthy {
				if probeStatus.Healthy {
					logrus.Infof("Probe [%s] is healthy", probeName)
				} else {
					logrus.Infof("Probe [%s] is unhealthy", probeName)
				}
			}
		}

		if allGood {
			logrus.Info("All probes are healthy")
			break
		}

		probeStatuses = newProbeStatuses
		initial = false
		time.Sleep(interval)
	}

	return nil
}
