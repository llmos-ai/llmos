package install

import (
	"fmt"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/pterm/pterm"

	"github.com/oneblock-ai/llmos/pkg/config"
	"github.com/oneblock-ai/llmos/pkg/questions"
)

const userPlaceHolder = "github:user1,github:user2"

// AskInstallDevice asks the user to choose the installation disk
func AskInstallDevice(cfg *config.LLMOSConfig) (*ghw.Disk, error) {
	var defaultDevice = "auto"
	var defaultDisk = &ghw.Disk{}
	maxSize := float64(0)

	disks := make(map[string]string)
	block, err := ghw.Block()
	if err == nil {
		for _, disk := range block.Disks {
			// skip useless devices (/dev/ram, /dev/loop, /dev/sr, /dev/zram)
			if strings.HasPrefix(disk.Name, "loop") || strings.HasPrefix(disk.Name, "ram") || strings.HasPrefix(disk.Name, "sr") || strings.HasPrefix(disk.Name, "zram") {
				continue
			}
			diskName := fmt.Sprintf("/dev/%s", disk.Name)
			size := float64(disk.SizeBytes) / float64(GiB)
			if size > maxSize {
				maxSize = size
				defaultDevice = diskName
				defaultDisk = disk
			}
			diskInfo := fmt.Sprintf("%s: %s(%.2f GiB) ", diskName, disk.Model, float64(disk.SizeBytes)/float64(GiB))
			disks[diskName] = diskInfo
		}
	}

	pterm.Info.Println("Available Disks:")
	for _, d := range disks {
		pterm.Info.Println(d)
	}

	device, err := questions.Prompt("Choose the installation disk:", defaultDevice, "Cannot be empty", false, false)
	if err != nil {
		return nil, err
	}

	if disks[device] == "" {
		return nil, fmt.Errorf("%s: %s", invalidDeviceNameError, device)
	}

	cfg.Install.Device = device

	if err = AskDataDevice(cfg, disks, device); err != nil {
		return nil, err
	}

	return defaultDisk, nil
}

// AskDataDevice asks the user to choose the data disk
func AskDataDevice(cfg *config.LLMOSConfig, devices map[string]string, rootDevice string) error {
	prompt := fmt.Sprintf("Use the installation disk(%s)", rootDevice)
	dataDevice, err := questions.Prompt("Choose the data disk:", rootDevice, prompt, true, false)
	if err != nil {
		return err
	}

	if devices[dataDevice] == "" {
		return fmt.Errorf("%s: %s", invalidDeviceNameError, dataDevice)
	}

	if cfg.Install.Device != dataDevice {
		cfg.Install.DataDevice = dataDevice
	}

	return nil
}

// AskConfigURL asks the user to provide the LLMOS config file location
func AskConfigURL(cfg *config.LLMOSConfig) error {
	if cfg.Install.ConfigURL != "" {
		return nil
	}

	str, err := questions.Prompt("LLMOS config file location (file path or http URL): ", "", "", true, false)
	if err != nil {
		return err
	}

	cfg.Install.ConfigURL = str
	return nil
}

// AskUserConfigs asks the user to provide the user accounting configurations
func AskUserConfigs(cfg *config.LLMOSConfig) error {
	if len(cfg.OS.SSHAuthorizedKeys) > 0 || cfg.OS.Password != "" {
		return nil
	}

	username, err := questions.Prompt("User to setup:", defaultLoginUser, emptyPlaceHolder, false, false)
	if err != nil {
		return err
	}

	passwd, err := questions.Prompt("Password:", "", emptyPlaceHolder, false, true)
	if err != nil {
		return err
	}

	users, err := questions.Prompt("SSH authorized keys(optional):", userPlaceHolder, emptyPlaceHolder, true, false)
	if err != nil {
		return err
	}

	// Cleanup the users if we selected the default values as they are not valid users
	if users == userPlaceHolder {
		users = ""
	}

	cfg.OS.Username = username
	cfg.OS.Password = passwd
	cfg.OS.SSHAuthorizedKeys = strings.Split(users, ",")

	return nil
}

func isYes(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "y") {
		return true
	}
	return false
}

func detectInstallationDevice() string {
	var device string
	maxSize := float64(0)

	block, err := ghw.Block()
	if err == nil {
		for _, disk := range block.Disks {
			size := float64(disk.SizeBytes) / float64(GiB)
			if size > maxSize {
				maxSize = size
				device = "/dev/" + disk.Name
			}
		}
	}
	return device
}
