package cmd

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type Runner interface {
	CmdExist(command string) bool
	InitCmd(string, ...string) *exec.Cmd
	RunCmd(cmd *exec.Cmd) ([]byte, error)
	Run(string, ...string) ([]byte, error)
}

type runner struct{}

func NewRunner() Runner {
	return &runner{}
}

func (r runner) CmdExist(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (r runner) InitCmd(command string, args ...string) *exec.Cmd {
	return exec.Command(command, args...)
}

func (r runner) RunCmd(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (r runner) Run(command string, args ...string) ([]byte, error) {
	slog.Debug(fmt.Sprintf("Running command: %s %s", command, strings.Join(args, " ")))
	cmd := r.InitCmd(command, args...)
	out, err := r.RunCmd(cmd)
	if err != nil {
		slog.Debug("command reported an error", "error", err.Error(), "output", string(out))
	}
	return out, err
}
