package install

import (
	"fmt"

	"github.com/jaypipes/ghw"

	"github.com/llmos-ai/llmos/pkg/config"
)

// Sanitize checks the install pre-conditions
func Sanitize(i config.Install) error {
	// Check if the target device has mounted partitions
	block, err := ghw.Block()
	if err == nil {
		for _, disk := range block.Disks {
			diskName := fmt.Sprintf("/dev/%s", disk.Name)
			if diskName == i.Device || diskName == i.DataDevice {
				for _, p := range disk.Partitions {
					if p.MountPoint != "" {
						return fmt.Errorf("target device %s has mounted partitions, please unmount them before installing", i.Device)
					}
				}

			}
		}
	}
	return nil
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
