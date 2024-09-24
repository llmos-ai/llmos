package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/plan"
	"github.com/llmos-ai/llmos/pkg/bootstrap/version"
	"github.com/llmos-ai/llmos/pkg/system"
	cliversion "github.com/llmos-ai/llmos/pkg/version"
)

type Config struct {
	Force             bool
	DataDir           string
	ConfigPath        string
	Token             string
	Server            string
	ClusterInit       bool
	Role              string
	KubernetesVersion string
}

// LLMOS is the main entrypoint to the llmos systemd service
type LLMOS struct {
	cfg Config
}

func New(cfg Config) *LLMOS {
	return &LLMOS{
		cfg: cfg,
	}
}
func (l *LLMOS) Run(ctx context.Context) error {
	if done, err := l.done(); err != nil {
		return fmt.Errorf("checking done stamp [%s]: %w", l.DoneStamp(), err)
	} else if done {
		logrus.Infof("System is already bootstrapped, " +
			"run with the --force flag to force bootstrap the system again.")
		return nil
	}

	for {
		err := l.execute(ctx)
		if err == nil {
			return nil
		}
		logrus.Warnf("failed to bootstrap system, will retry: %v", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
		}
	}
}

func (l *LLMOS) execute(ctx context.Context) error {
	cfg, err := config.Load(l.cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg = mergeConfigs(l.cfg, cfg)

	if err = validateConfig(&cfg); err != nil {
		// terminate bootstrap if config is invalid
		logrus.Fatalf("invalid config: %v", err)
	}

	if err = l.setWorking(cfg); err != nil {
		return fmt.Errorf("failed to save working config to %s: %w", l.WorkingStamp(), err)
	}

	var k8sVersion, operatorVersion string
	if cfg.Role == config.ClusterInitRole {
		k8sVersion, err = version.K8sVersion(cfg.KubernetesVersion)
		if err != nil {
			return err
		}

		operatorVersion, err = version.OperatorVersion(cfg.ChartRepo, cfg.LLMOSOperatorVersion)
		if err != nil {
			return err
		}
	} else {
		k8sVersion, operatorVersion, err = version.GetClusterK8sAndOperatorVersions(cfg.Server, cfg.Token)
		if err != nil {
			return err
		}
		cfg.KubernetesVersion = k8sVersion
		cfg.LLMOSOperatorVersion = operatorVersion
	}

	logrus.Infof("Bootstrapping LLMOS %s(%s)", operatorVersion, k8sVersion)

	nodePlan, err := plan.ToPlan(ctx, &cfg, l.cfg.DataDir)
	if err != nil {
		return fmt.Errorf("generating plan: %w", err)
	}
	logrus.Debugf("Generated node plan: %+v", nodePlan)

	if err = plan.Run(ctx, &cfg, nodePlan, l.cfg.DataDir); err != nil {
		return fmt.Errorf("running plan error: %w", err)
	}

	if err = l.setDone(cfg); err != nil {
		return err
	}

	logrus.Infof("Successfully Bootstrapped LLMOS %s(%s)", operatorVersion, k8sVersion)
	return nil
}

func (l *LLMOS) writeConfig(path string, cfg config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0600); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

func (l *LLMOS) setWorking(cfg config.Config) error {
	return l.writeConfig(l.WorkingStamp(), cfg)
}

func (l *LLMOS) setDone(cfg config.Config) error {
	return l.writeConfig(l.DoneStamp(), cfg)
}

func (l *LLMOS) done() (bool, error) {
	if l.cfg.Force {
		_ = os.Remove(l.DoneStamp())
		return false, nil
	}
	_, err := os.Stat(l.DoneStamp())
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LLMOS) DoneStamp() string {
	return system.DataPath("bootstrapped")
}

func (l *LLMOS) WorkingStamp() string {
	return filepath.Join(l.cfg.DataDir, "working")
}

func (l *LLMOS) Info(ctx context.Context) error {
	operatorVersion, k8sVersion, osVersion := l.getExistingVersions(ctx)
	fmt.Printf(" OS Version: 	 %s\n", osVersion)
	fmt.Printf(" LLMOS Operator: %s\n", operatorVersion)
	fmt.Printf(" LLMOS Cli:	 %s\n", cliversion.GetFriendlyVersion())
	fmt.Printf(" Kubernetes:	 %s\n\n", k8sVersion)
	return nil
}
