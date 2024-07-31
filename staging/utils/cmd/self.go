package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
)

func Self() (string, error) {
	cmd := os.Args[0]
	if _, err := os.Stat(cmd); err == nil {
		return filepath.Abs(cmd)
	}
	cmd, err := exec.LookPath(cmd)
	if err != nil {
		return "", err
	}
	return filepath.Abs(cmd)
}
