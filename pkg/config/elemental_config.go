package config

import (
	"fmt"
	"path/filepath"

	"github.com/jaypipes/ghw"
	"github.com/rancher/elemental-toolkit/pkg/constants"
)

const (
	SoftMinDiskSizeGiB   = 140
	HardMinDiskSizeGiB   = 60
	MinCosPartSizeGiB    = 25
	NormalCosPartSizeGiB = 50
)

type ElementalConfig struct {
	Install ElementalInstallSpec `yaml:"install,omitempty"`
}

type ElementalInstallSpec struct {
	Target          string                     `yaml:"target,omitempty"`
	Partitions      *ElementalDefaultPartition `yaml:"partitions,omitempty"`
	ExtraPartitions []ElementalPartition       `yaml:"extra-partitions,omitempty"`
	CloudInit       string                     `yaml:"cloud-init,omitempty"`
	System          *ElementalSystem           `yaml:"system,omitempty"`
	TTY             string                     `yaml:"tty,omitempty"`
}

type ElementalSystem struct {
	Label string `yaml:"label,omitempty"`
	Size  uint   `yaml:"size,omitempty"`
	FS    string `yaml:"fs,omitempty"`
	URI   string `yaml:"uri,omitempty"`
}

type ElementalDefaultPartition struct {
	OEM        *ElementalPartition `yaml:"oem,omitempty"`
	State      *ElementalPartition `yaml:"state,omitempty"`
	Recovery   *ElementalPartition `yaml:"recovery,omitempty"`
	Persistent *ElementalPartition `yaml:"persistent,omitempty"`
}

type ElementalPartition struct {
	FilesystemLabel string `yaml:"label,omitempty"`
	Size            uint   `yaml:"size,omitempty"`
	FS              string `yaml:"fs,omitempty"`
}

func NewElementalConfig(path, configUrl, tty string) *ElementalConfig {
	return &ElementalConfig{
		Install: ElementalInstallSpec{
			Target:    path,
			CloudInit: configUrl,
			TTY:       tty,
		},
	}
}

func (c *LLMOSConfig) HasDataPartition() bool {
	if c.Install.DataDevice == "" {
		return false
	}

	return true
}

func GenerateElementalConfig(cfg *LLMOSConfig, rootDisk *ghw.Disk) (*ElementalConfig, error) {
	path, err := filepath.EvalSymlinks(cfg.Install.Device)
	if err != nil {
		return nil, err
	}
	elementalConfig := NewElementalConfig(path, cfg.Install.ConfigURL, cfg.Install.TTY)

	//customize data partition layout
	elementalConfig, err = CreateRootPartitioningLayout(cfg, elementalConfig, rootDisk)
	if err != nil {
		return nil, err
	}

	return elementalConfig, nil
}

func CreateRootPartitioningLayout(cfg *LLMOSConfig, elementalConfig *ElementalConfig, rootDisk *ghw.Disk) (*ElementalConfig, error) {
	var err error
	cosPersistentSizeGiB := uint64(0)
	if cfg.HasDataPartition() {
		diskSizeBytes := rootDisk.SizeBytes
		cosPersistentSizeGiB, err = calcCosPersistentPartSize(diskSizeBytes >> 30)
		if err != nil {
			return nil, err
		}
	}

	elementalConfig.Install.Partitions = &ElementalDefaultPartition{
		OEM: &ElementalPartition{
			FilesystemLabel: constants.OEMLabel,
			Size:            constants.OEMSize,
			FS:              constants.LinuxFs,
		},
		State: &ElementalPartition{
			FilesystemLabel: constants.StateLabel,
			Size:            constants.StateSize,
			FS:              constants.LinuxFs,
		},
		Recovery: &ElementalPartition{
			FilesystemLabel: constants.RecoveryLabel,
			Size:            constants.RecoverySize,
			FS:              constants.LinuxFs,
		},
		Persistent: &ElementalPartition{
			FilesystemLabel: constants.PersistentLabel,
			Size:            uint(cosPersistentSizeGiB << 10),
			FS:              constants.LinuxFs,
		},
	}

	if cfg.HasDataPartition() {
		elementalConfig.Install.ExtraPartitions = []ElementalPartition{
			{
				FilesystemLabel: "LLMOS_DATA_PERSISTENT",
				Size:            0,
				FS:              "ext4",
			},
		}
	}

	return elementalConfig, nil
}

func calcCosPersistentPartSize(diskSizeGiB uint64) (uint64, error) {
	switch {
	case diskSizeGiB < HardMinDiskSizeGiB:
		return 0, fmt.Errorf("disk too small: %dGB. Minimum %dGB is required", diskSizeGiB, HardMinDiskSizeGiB)
	case diskSizeGiB < SoftMinDiskSizeGiB:
		d := MinCosPartSizeGiB / float64(SoftMinDiskSizeGiB-HardMinDiskSizeGiB)
		partSizeGiB := MinCosPartSizeGiB + float64(diskSizeGiB-HardMinDiskSizeGiB)*d
		return uint64(partSizeGiB), nil
	default:
		partSizeGiB := NormalCosPartSizeGiB + ((diskSizeGiB-100)/100)*10
		if partSizeGiB > 100 {
			partSizeGiB = 100
		}
		return partSizeGiB, nil
	}
}
