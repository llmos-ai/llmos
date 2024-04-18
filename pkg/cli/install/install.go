package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils"
	"github.com/llmos-ai/llmos/pkg/utils/cmd"
	"github.com/llmos-ai/llmos/pkg/utils/log"
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
	invalidDeviceNameError = "invalid device name"
)

type Installer struct {
	LLMOSConfig  *config.Config
	runner       cmd.Runner
	logger       log.Logger
	elementalCli elemental.Elemental
}

func NewInstaller(cfg *config.Config, logger log.Logger) *Installer {
	return &Installer{
		LLMOSConfig:  cfg,
		logger:       logger,
		elementalCli: elemental.NewElemental(),
		runner:       cmd.NewRunner(),
	}
}

func (i *Installer) RunInstall() error {
	if i.LLMOSConfig.Install.Device == "" || i.LLMOSConfig.Install.Device == "auto" {
		i.LLMOSConfig.Install.Device = detectInstallationDevice()
	}

	if i.LLMOSConfig.Install.Device == "" {
		return fmt.Errorf("no device found to install LLMOS")
	}

	return i.GenerateInstallConfigs()
}

func (i *Installer) runInstall() error {
	i.logger.Info("Running install")
	utils.SetEnv(i.LLMOSConfig.Install.Env)

	if err := Sanitize(i.LLMOSConfig.Install); err != nil {
		return err
	}

	if err := i.elementalCli.Install(i.LLMOSConfig.Install); err != nil {
		return err
	}

	i.logger.Info("Installation complete")
	return nil
}

func (i *Installer) GenerateInstallConfigs() error {
	var configUrls []string

	// add llmos config file
	llmOSConfigFile, err := utils.SaveTemp(i.LLMOSConfig, "llmos-config", i.logger, true)
	if err != nil {
		return err
	}
	defer os.Remove(llmOSConfigFile)

	// add after install chroot files
	afterInstallStage, err := config.AddStageAfterInstallChroot(llmOSConfigFile, i.LLMOSConfig)
	if err != nil {
		return err
	}

	// add cos config file
	cosConfig, err := config.ConvertToCosStages(i.LLMOSConfig, *afterInstallStage)
	if err != nil {
		return err
	}

	cosConfigFile, err := utils.SaveTemp(cosConfig, "cos-config", i.logger, false)
	if err != nil {
		return err
	}
	defer os.Remove(cosConfigFile)
	configUrls = append(configUrls, cosConfigFile)

	// add the cosConfig file to the cloud-init config files of install
	i.LLMOSConfig.ConfigURL = strings.Join(configUrls[:], ",")

	// add elemental config dir and file
	elementalConfig, err := elemental.GenerateElementalConfig(i.LLMOSConfig)
	if err != nil {
		return err
	}

	elementalConfigDir, elementalConfigFile, err := utils.SaveElementalConfig(elementalConfig, i.logger)
	if err != nil {
		return err
	}
	defer os.Remove(elementalConfigFile)

	// specify the elemental install config-dir
	i.LLMOSConfig.Install.ConfigDir = elementalConfigDir

	return i.runInstall()
}
