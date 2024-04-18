package elemental

import (
	"path/filepath"

	elconst "github.com/rancher/elemental-toolkit/pkg/constants"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/constants"
)

type ElementalConfig struct {
	Install  InstallSpec `yaml:"install,omitempty"`
	Reboot   bool        `yaml:"reboot,omitempty"`
	Poweroff bool        `yaml:"poweroff,omitempty"`
}

type InstallSpec struct {
	Target          string            `yaml:"target,omitempty"`
	Partitions      *DefaultPartition `yaml:"partitions,omitempty"`
	ExtraPartitions []Partition       `yaml:"extra-partitions,omitempty"`
	ISO             string            `yaml:"iso,omitempty"`
	CloudInit       string            `yaml:"cloud-init,omitempty"`
	System          string            `yaml:"system,omitempty"`
	TTY             string            `yaml:"tty,omitempty"`
}

type DefaultPartition struct {
	OEM        *Partition `yaml:"oem,omitempty"`
	State      *Partition `yaml:"state,omitempty"`
	Recovery   *Partition `yaml:"recovery,omitempty"`
	Persistent *Partition `yaml:"persistent,omitempty"`
}

type Partition struct {
	FilesystemLabel string `yaml:"label,omitempty"`
	Size            uint   `yaml:"size,omitempty"`
	FS              string `yaml:"fs,omitempty"`
}

func NewElementalConfig(path string, i config.Install, configURL string) *ElementalConfig {
	cfg := &ElementalConfig{
		Install: InstallSpec{
			Target: path,
		},
		Reboot:   i.Reboot,
		Poweroff: i.PowerOff,
	}

	if configURL != "" {
		cfg.Install.CloudInit = configURL
	}

	if i.TTY != "" {
		cfg.Install.TTY = i.TTY
	}

	if i.SystemURI != "" {
		cfg.Install.System = i.SystemURI
	}

	if i.ISOUrl != "" {
		cfg.Install.ISO = i.ISOUrl
	}

	return cfg
}

func GenerateElementalConfig(cfg *config.Config) (*ElementalConfig, error) {
	path, err := filepath.EvalSymlinks(cfg.Install.Device)
	if err != nil {
		return nil, err
	}
	elementalConfig := NewElementalConfig(path, cfg.Install, cfg.ConfigURL)

	//customize data partition layout
	elementalConfig, err = CreateRootPartitioningLayout(cfg, elementalConfig)
	if err != nil {
		return nil, err
	}

	return elementalConfig, nil
}

func CreateRootPartitioningLayout(cfg *config.Config, elementalConfig *ElementalConfig) (*ElementalConfig, error) {
	elementalConfig.Install.Partitions = &DefaultPartition{
		OEM: &Partition{
			FilesystemLabel: elconst.OEMLabel,
			Size:            elconst.OEMSize,
			FS:              elconst.LinuxFs,
		},
		State: &Partition{
			FilesystemLabel: elconst.StateLabel,
			Size:            constants.StateSize, // adding more size for air-gap images
			FS:              elconst.LinuxFs,
		},
		Recovery: &Partition{
			FilesystemLabel: elconst.RecoveryLabel,
			Size:            constants.RecoverySize, // ditto
			FS:              elconst.LinuxFs,
		},
		Persistent: &Partition{
			FilesystemLabel: elconst.PersistentLabel,
			Size:            elconst.PersistentSize,
			FS:              elconst.LinuxFs,
		},
	}

	if cfg.HasDataPartition() {
		elementalConfig.Install.ExtraPartitions = []Partition{
			{
				FilesystemLabel: "LLMOS_DATA_PERSISTENT",
				Size:            elconst.PersistentSize,
				FS:              elconst.LinuxFs,
			},
		}
	}

	return elementalConfig, nil
}
