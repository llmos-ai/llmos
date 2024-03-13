package install

import (
	"fmt"

	"github.com/guangbochen/golib/disk"
	"github.com/jaypipes/ghw"

	"github.com/oneblock-ai/llmos/pkg/config"
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

func formatDisk(path string) error {
	return disk.MakeExt4DiskFormatting(path, "")
}
