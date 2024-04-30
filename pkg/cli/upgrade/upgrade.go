package upgrade

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pterm/pterm"
	"k8s.io/utils/exec"
	"k8s.io/utils/nsenter"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/system"
	"github.com/llmos-ai/llmos/pkg/utils"
	"github.com/llmos-ai/llmos/pkg/utils/log"
)

const (
	osReleaseFile  = "/etc/os-release"
	systemStopping = "stopping"
	cosStaetFile   = "/run/initramfs/elemental-state/state.yaml"

	osImagePrefix    = "IMAGE="
	osImageTagPrefix = "IMAGE_TAG="
)

var (
	hostMountDirs = []string{
		"/dev",
		"/run",
	}
)

type Upgrade struct {
	Config    config.Upgrade
	logger    log.Logger
	elemental elemental.Elemental
	ns        *nsenter.NSEnter
	osRelease string
}

func NewUpgrade(logger log.Logger, cfg config.Upgrade) (*Upgrade, error) {
	ns, err := nsenter.NewNsenter(cfg.HostDir, exec.New())
	if err != nil {
		return nil, fmt.Errorf("failed to create nsenter with %s: %v", cfg.HostDir, err)
	}

	return &Upgrade{
		Config:    cfg,
		logger:    logger,
		elemental: elemental.NewElemental(),
		ns:        ns,
	}, nil
}

func (u *Upgrade) Run() error {
	var err error
	if !utils.IsK8sPod() && !u.Config.Dev {
		return fmt.Errorf("upgrade only support run in Kubernetes pod now")
	}

	if _, err = u.checkSystemStatus(); err != nil {
		return err
	}

	newVersion, err := u.hasNewerOSVersion()
	if err != nil {
		// allow dev upgrade even if the version is invalid, e.g., main, dev
		if !u.Config.Dev && !errors.Is(err, semver.ErrInvalidSemVer) {
			return err
		}
	}

	if !newVersion && !u.Config.Force {
		return fmt.Errorf("image OS version is not newer than the current version, use `--force` flag to run a force upgrade")
	}

	if err = u.mountHostDirs(); err != nil {
		return err
	}

	u.logger.Info("Upgrading to new OS version")
	pterm.Info.Println(u.osRelease)
	if err = u.elemental.Upgrade(u.Config); err != nil {
		return fmt.Errorf("failed to upgrade: %v", err)
	}

	u.logger.Info("Upgrade complete, rebooting the system now")

	return u.ns.Command("reboot").Run()
}

func (u *Upgrade) checkSystemStatus() (string, error) {
	res, err := u.ns.Command("systemctl", "is-system-running").Output()
	if err != nil {
		return "", fmt.Errorf("failed to check system status: %v", err)
	}

	status := strings.Trim(string(res), "\n")

	if status == systemStopping {
		return "", fmt.Errorf("system is stopping, cannot upgrade now")
	}

	u.logger.Debug("Host system status", "status", status)
	return status, nil
}

func (u *Upgrade) hasNewerOSVersion() (bool, error) {
	osRelease, err := os.ReadFile(osReleaseFile)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %v", osRelease, err)
	}
	u.osRelease = string(osRelease)

	hostOSRelease, err := os.ReadFile(system.HostRootPath(osReleaseFile))
	if err != nil {
		return false, fmt.Errorf("failed to read host OS release: %v", err)
	}

	if string(osRelease) == string(hostOSRelease) {
		u.logger.Info("OS version is the same as the host OS version, skipping upgrade")
		pterm.Info.Print(string(osRelease))
		return false, nil
	}

	oslines := strings.Split(string(osRelease), "\n")
	hostLines := strings.Split(string(hostOSRelease), "\n")
	return compareVersion(oslines, hostLines)
}

func (u *Upgrade) mountHostDirs() error {
	for _, dir := range hostMountDirs {
		// check if mount dir exist in the host
		hostDir := system.HostRootPath(dir)
		if _, err := os.Stat(hostDir); err != nil {
			return fmt.Errorf("failed to stat host mount dir %s: %v", dir, err)
		}

		if err := exec.New().Command("mount", "--rbind", hostDir, dir).Run(); err != nil {
			return fmt.Errorf("failed to mount host dir %s: %v", dir, err)
		}
	}

	stateFile, err := os.ReadFile(cosStaetFile)
	if err != nil {
		return fmt.Errorf("failed to read state config file: %v", err)
	}
	if u.logger.IsDebug() {
		u.logger.Info("read state file:", "path", cosStaetFile)
		pterm.Info.Print(string(stateFile))
	}
	return nil
}
