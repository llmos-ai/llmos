package plan

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/kubectl"
	operator "github.com/llmos-ai/llmos/pkg/bootstrap/llmos-operator"
	"github.com/llmos-ai/llmos/pkg/bootstrap/manifest"
	"github.com/llmos-ai/llmos/pkg/bootstrap/registry"
	"github.com/llmos-ai/llmos/pkg/bootstrap/role"
	"github.com/llmos-ai/llmos/pkg/bootstrap/runtime"
	"github.com/llmos-ai/llmos/pkg/bootstrap/version"
	"github.com/llmos-ai/llmos/pkg/cli/probe"
)

type plan applyinator.Plan

func toInitPlan(cfg *config.Config, dataDir string) (*applyinator.Plan, error) {
	logrus.Info("generating init plan")
	if err := assignTokenIfUnset(cfg); err != nil {
		return nil, err
	}

	var p = plan{}
	if err := p.addFiles(cfg, dataDir); err != nil {
		return nil, err
	}

	if err := p.addInstructions(cfg, dataDir, true); err != nil {
		return nil, err
	}

	if err := p.addProbes(cfg); err != nil {
		return nil, err
	}

	return (*applyinator.Plan)(&p), nil
}

func toJoinPlan(cfg *config.Config, dataDir string) (*applyinator.Plan, error) {
	var (
		cfgRole      = string(cfg.Role)
		etcd         = role.IsEtcd(cfgRole)
		controlPlane = role.IsControlPlane(cfgRole)
		worker       = role.IsWorker(cfgRole)
	)

	if !etcd && !controlPlane && !worker {
		return nil, fmt.Errorf("invalid role (%s) defined", cfgRole)
	}
	if cfg.Server == "" {
		return nil, fmt.Errorf("server is required in config for all roles besides cluster-init")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("token is required in config for all roles besides cluster-init")
	}

	logrus.Info("generating join plan")
	p := plan{}
	// add join files
	if err := p.addJoinFiles(cfg, dataDir); err != nil {
		return nil, err
	}

	// add join instructions
	if err := p.addInstructions(cfg, dataDir, false); err != nil {
		return nil, err
	}

	// add probe instruction
	if err := p.addInstruction(probe.ToInstruction()); err != nil {
		return nil, err
	}

	// add join probes
	p.addProbesForJoin(cfg)

	return (*applyinator.Plan)(&p), nil
}

func ToPlan(_ context.Context, cfg *config.Config, dataDir string) (*applyinator.Plan, error) {
	newCfg := *cfg
	if newCfg.Role == config.ClusterInitRole {
		return toInitPlan(&newCfg, dataDir)
	}

	return toJoinPlan(&newCfg, dataDir)
}

func (p *plan) addInstructions(cfg *config.Config, dataDir string, initRole bool) error {
	k8sVersion, err := version.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}

	// add k8s runtime instruction
	if err = p.addInstruction(runtime.ToInstruction(cfg, k8sVersion)); err != nil {
		return err
	}

	// add probe instruction
	if err = p.addInstruction(probe.ToInstruction()); err != nil {
		return err
	}

	if cfg.Role != config.AgentRole {
		// Copy kubeconfig for cluster-init and server node
		if err = p.addInstruction(runtime.CopyKubeConfigInstruction(k8sVersion)); err != nil {
			return err
		}
	}

	// only need to apply operator charts & resource on cluster-init role
	if initRole {
		operatorVersion, err := version.OperatorVersion(cfg.ChartRepo, cfg.LLMOSOperatorVersion)
		if err != nil {
			return err
		}

		// add resource instruction
		if err := p.addInstruction(manifest.ToInstruction(k8sVersion, manifest.GetBootstrapManifests(dataDir))); err != nil {
			return err
		}

		// add operator chart config instruction
		if err := p.addInstruction(operator.ToChartConfigInstruction(k8sVersion, dataDir)); err != nil {
			return err
		}

		if err := p.addInstruction(operator.ToInstruction(cfg.LLMOSInstallerImage,
			cfg.GlobalSystemImageRegistry, cfg.Mirror, k8sVersion, operatorVersion)); err != nil {
			return err
		}

		// add operator wait instructions
		if err := p.addInstruction(operator.ToWaitOperatorInstruction(k8sVersion)); err != nil {
			return err
		}

		if err := p.addInstruction(operator.ToWaitOperatorWebhookInstruction(k8sVersion)); err != nil {
			return err
		}

		if err := p.addInstruction(manifest.ToInstruction(k8sVersion,
			manifest.GetBootstrapPrePostManifests(dataDir))); err != nil {
			return err
		}

		if err := p.addInstruction(operator.ToWaitSUCInstruction(cfg.LLMOSInstallerImage,
			cfg.GlobalSystemImageRegistry, k8sVersion)); err != nil {
			return err
		}
	}

	// Add wait instructions
	if cfg.Role == config.AgentRole {
		if err = p.addInstruction(runtime.ToWaitSystemAgentActiveInstruction(k8sVersion)); err != nil {
			return err
		}
	} else {
		nodeName, err := manifest.GetNodeName(cfg)
		if err != nil {
			return err
		}
		if err = p.addInstruction(runtime.ToWaitNodeReadyInstruction(nodeName, k8sVersion)); err != nil {
			return err
		}
	}

	p.addPrePostInstructions(cfg, k8sVersion)
	return nil
}

func (p *plan) addPrePostInstructions(cfg *config.Config, k8sVersion string) {
	var instructions = make([]applyinator.OneTimeInstruction, 0)

	for _, inst := range cfg.PreInstructions {
		if k8sVersion != "" {
			inst.Env = append(inst.Env, kubectl.Env(k8sVersion)...)
		}
		instructions = append(instructions, inst)
	}

	instructions = append(instructions, p.OneTimeInstructions...)

	for _, inst := range cfg.PostInstructions {
		inst.Env = append(inst.Env, kubectl.Env(k8sVersion)...)
		instructions = append(instructions, inst)
	}

	p.OneTimeInstructions = instructions
}

func (p *plan) addInstruction(instruction *applyinator.OneTimeInstruction, err error) error {
	if err != nil || instruction == nil {
		return err
	}

	p.OneTimeInstructions = append(p.OneTimeInstructions, *instruction)
	return nil
}

func (p *plan) addFiles(cfg *config.Config, dataDir string) error {
	k8sVersions, err := version.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}

	runtimeName := config.GetRuntime(k8sVersions)
	if runtimeName == config.RuntimeUnknown {
		return fmt.Errorf("unknown runtime %s", runtimeName)
	}

	// bootstrap config.yaml
	if err = p.addFile(runtime.ToBootstrapFile(&cfg.RuntimeConfig, runtimeName, cfg.Server)); err != nil {
		return err
	}

	// add pre-post manifests
	if err = p.addFile(manifest.ToBootstrapPrePostFile(cfg,
		manifest.GetBootstrapPrePostManifests(dataDir))); err != nil {
		return err
	}

	// add token file
	if err = p.addFile(runtime.ToTokenFile(cfg.Token, dataDir)); err != nil {
		return err
	}

	// registries.yaml
	if err = p.addFile(registry.ToFile(cfg.Registries, runtimeName)); err != nil {
		return err
	}

	// bootstrap manifests
	if err = p.addFile(manifest.ToBootstrapFile(cfg,
		manifest.GetBootstrapManifests(dataDir), runtimeName)); err != nil {
		return err
	}

	// llmos operator values.yaml
	return p.addFile(operator.ToFile(cfg, dataDir))
}

func (p *plan) addJoinFiles(cfg *config.Config, dataDir string) error {
	k8sVersions, err := version.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	runtimeName := config.GetRuntime(k8sVersions)

	// config.yaml
	if err = p.addFile(runtime.ToBootstrapFile(&cfg.RuntimeConfig, runtimeName, cfg.Server)); err != nil {
		return err
	}

	// add token file
	if err = p.addFile(runtime.ToTokenFile(cfg.Token, dataDir)); err != nil {
		return err
	}

	return nil
}

func (p *plan) addFile(file *applyinator.File, err error) error {
	if err != nil || file == nil {
		return err
	}
	p.Files = append(p.Files, *file)
	return nil
}

func (p *plan) addProbesForJoin(cfg *config.Config) {
	p.Probes = probe.ProbesForJoin(&cfg.RuntimeConfig)
}

func (p *plan) addProbes(cfg *config.Config) error {
	k8sVersion, err := version.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	p.Probes = probe.AllProbes(config.GetRuntime(k8sVersion))
	return nil
}
