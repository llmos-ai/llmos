package upgrade

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Masterminds/semver"
)

func compareVersion(osLines, hostLines []string) (bool, error) {
	if len(osLines) == 0 || len(hostLines) == 0 {
		return false, fmt.Errorf("either image or host OS release file is empty")
	}
	var (
		imageFound, imageTagFound bool
		osImage                   string
		osImageTag, hostImageTag  *semver.Version
		err                       error
	)

	for i := 0; i < len(osLines) || i < len(hostLines); i++ {
		if i < len(osLines) {
			l := osLines[i]
			if strings.HasPrefix(l, osImagePrefix) {
				imageFound = true
				osImage = l
			}

			if strings.HasPrefix(l, osImageTagPrefix) {
				imageTagFound = true
				osImageTag, err = getTagVersion(l)
				if err != nil {
					return false, fmt.Errorf("failed to parse version from %s: %v", l, err)
				}
			}
		}

		if i < len(hostLines) {
			l2 := hostLines[i]
			if strings.HasPrefix(l2, osImagePrefix) {
				if !imageFound {
					osImage = l2
					imageFound = true
				} else if osImage != l2 {
					return false, fmt.Errorf("OS image is different: %s != %s, use `--force` flag to run a force upgrade", osImage, l2)
				}
			}

			if strings.HasPrefix(l2, osImageTagPrefix) {
				imageTagFound = true
				hostImageTag, err = getTagVersion(l2)
				if err != nil {
					return false, fmt.Errorf("failed to parse version from %s: %v", l2, err)
				}
			}
		}

		if imageFound && imageTagFound && osImageTag != nil && hostImageTag != nil {
			slog.Debug("compare OS image tags", "OS IMAGE_TAG", osImageTag.String(), "Host IMAGE_TAG", hostImageTag.String())
			if osImageTag.GreaterThan(hostImageTag) {
				return true, nil
			}
			return false, fmt.Errorf("current OS image tag is either older or equal to the host OS image tag, %s <= %s", osImageTag, hostImageTag)
		}
	}

	return false, fmt.Errorf("failed to compare OS release, either IMAGE:%t or IMAGE_TAG:%t not found", imageFound, imageTagFound)
}

func getTagVersion(val string) (*semver.Version, error) {
	parts := strings.Split(val, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to parse version from %s", val)
	}
	version := strings.Trim(parts[1], "\"")
	return semver.NewVersion(version)
}
