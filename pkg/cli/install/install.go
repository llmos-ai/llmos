package install

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"

	"github.com/oneblock-ai/llmos/pkg/config"
	"github.com/oneblock-ai/llmos/pkg/questions"
	"github.com/oneblock-ai/llmos/pkg/utils"
)

const (
	_ = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
)

const (
	emptyPlaceHolder = "Unset"
	yesOrNo          = "[Y]es/[n]o"

	defaultLoginUser       = "llmos"
	defaultLogFilePath     = "/var/log/llmos-install.log"
	invalidDeviceNameError = "invalid device name"
	oemTargetPath          = "/run/cos/oem/"
)

type configFiles struct {
	elementalConfigDir  string
	elementalConfigFile string
	cosConfigFile       string
	llmOSConfigFile     string
}

func AskInstall(cfg *config.LLMOSConfig) error {
	if cfg.Install.Silent {
		return nil
	}

	pterm.Info.Println("Welcome to the LLMOS installer")
	rootDisk, err := AskInstallDevice(cfg)
	if err != nil {
		if strings.Contains(err.Error(), invalidDeviceNameError) {
			pterm.Error.Println(err.Error())
			return AskInstall(cfg)
		}
		return err
	}

	if err = AskConfigURL(cfg); err != nil {
		return err
	}

	if err = AskUserConfigs(cfg); err != nil {
		return err
	}

	allGood, err := questions.Prompt("Are settings ok?", "n", yesOrNo, true, false)
	if err != nil {
		return err
	}

	if !isYes(allGood) {
		return AskInstall(cfg)
	}

	cfs := configFiles{}
	cosConfig, err := ConvertToCos(cfg)
	if err != nil {
		return err
	}

	cfs.cosConfigFile, err = utils.SaveTemp(cosConfig, "cos")
	if err != nil {
		return err
	}
	defer os.Remove(cfs.cosConfigFile)

	cfs.llmOSConfigFile, err = utils.SaveTemp(cfg, "llmos")
	if err != nil {
		return err
	}
	defer os.Remove(cfs.llmOSConfigFile)

	cfg.Install.ConfigURL = cfs.cosConfigFile

	// create a tmp config file for installation
	elementalConfig, err := config.GenerateElementalConfig(cfg, rootDisk)
	if err != nil {
		return err
	}

	cfs.elementalConfigDir, cfs.elementalConfigFile, err = utils.SaveElementalConfig(elementalConfig)
	if err != nil {
		return err
	}
	defer os.Remove(cfs.elementalConfigFile)

	if err = RunInstall(cfg, cfs); err != nil {
		return err
	}

	return nil
}

func RunInstall(cfg *config.LLMOSConfig, cfs configFiles) error {
	utils.SetEnv(cfg.OS.Env)
	utils.SetEnv(cfg.Install.Env)

	if cfg.Install.Device == "" || cfg.Install.Device == "auto" {
		cfg.Install.Device = detectInstallationDevice()
	}

	if err := runInstall(cfg, cfs); err != nil {
		return err
	}

	// copy template file to the /oem config directory
	// Note: don't use yaml file extension for the config files as it will be applied again
	if err := utils.CopyFile(cfs.llmOSConfigFile, oemTargetPath+"llmos.config"); err != nil {
		return err
	}

	if err := utils.CopyFile(cfs.elementalConfigFile, oemTargetPath+"elemental.config"); err != nil {
		return err
	}

	return nil
}

func runInstall(cfg *config.LLMOSConfig, cfs configFiles) error {
	slog.Info("Running install")
	if err := Sanitize(cfg.Install); err != nil {
		return err
	}

	if err := DeactivateDevices(); err != nil {
		return err
	}

	// run the elemental install
	args := []string{
		"install", "--config-dir", cfs.elementalConfigDir,
		"--debug",
	}
	cmd := exec.Command("elemental", args...)
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	// Execute the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("elemental install failed: %s", err)
	}
	slog.Info(stdBuffer.String())

	pterm.Info.Println("Installation complete")
	return nil
}

// DeactivateDevices helps to tear down LVM and MD devices on the system, if the installing device is occupied, the partitioning operation could fail later.
func DeactivateDevices() error {
	slog.Info("Deactivating LVM and MD devices")
	cmd := exec.Command("blkdeactivate", "--lvmoptions", "wholevg,retry",
		"--dmoptions", "force,retry", "--errors")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deactivating LVM and MD devices failed: %s", err)
	}
	return nil
}
