package plan

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/llmos-ai/llmos/pkg/applyinator"
	"github.com/llmos-ai/llmos/pkg/applyinator/image"
	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
	"github.com/llmos-ai/llmos/pkg/bootstrap/versions"
)

const defaultInsAttempts = 3

func Run(ctx context.Context, cfg *config.Config, plan *applyinator.Plan, dataDir string) error {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return err
	}
	return RunWithKubernetesVersion(ctx, cfg, k8sVersion, plan, dataDir)
}

func RunWithKubernetesVersion(ctx context.Context, cfg *config.Config, k8sVersion string,
	plan *applyinator.Plan, dataDir string) error {
	logrus.Infof("Running plan for Kubernetes version %s, plan: %v, datadir: %s", k8sVersion, plan.OneTimeInstructions, dataDir)

	if err := writePlan(plan, dataDir); err != nil {
		return err
	}

	// init apply plan
	images := image.NewUtility(cfg.ImageUtility)
	apply := applyinator.NewApplyinator(filepath.Join(dataDir, "plan", "work"),
		false, filepath.Join(dataDir, "plan", "applied"), "", images)

	output, err := apply.Apply(ctx, applyinator.ApplyInput{
		CalculatedPlan: applyinator.CalculatedPlan{
			Plan: *plan,
		},
		RunOneTimeInstructions:     true,
		ReconcileFiles:             true,
		OneTimeInstructionAttempts: defaultInsAttempts,
	})

	if err != nil || !output.OneTimeApplySucceeded {
		return fmt.Errorf("failed to apply plan: %w", err)
	}

	return saveOutput(output.OneTimeOutput, dataDir)
}

func saveOutput(data []byte, dataDir string) error {
	planOutput := GetPlanOutput(dataDir)
	f, err := os.OpenFile(planOutput, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	in, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, in)
	return err
}

func writePlan(plan *applyinator.Plan, dataDir string) error {
	planFile := GetPlanFile(dataDir)
	if err := os.MkdirAll(filepath.Dir(planFile), 0755); err != nil {
		return err
	}

	logrus.Infof("Writing plan file to %s", planFile)
	f, err := os.OpenFile(planFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}

func GetPlanFile(dataDir string) string {
	return filepath.Join(dataDir, "plan", "plan.json")
}

func GetPlanOutput(dataDir string) string {
	return filepath.Join(dataDir, "plan", "plan-output.json")
}
