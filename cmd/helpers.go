package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/llmos-ai/llmos/pkg/constants"
	"github.com/llmos-ai/llmos/pkg/utils/log"
)

// CheckRoot helps to check if the user is root
func CheckRoot(devMode bool) error {
	if devMode {
		return nil
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges is required to run this command. Please run with sudo or as root user")
	}
	return nil
}

func CheckSource(source string) error {
	if source == "" {
		if !checkIsLiveMode() {
			return fmt.Errorf("source must be provided if is not live ISO")
		}
		return nil
	}

	r, err := regexp.Compile(`^oci:|dir:|file:`)
	if err != nil {
		return err
	}
	if !r.MatchString(source) {
		return fmt.Errorf("source must be one of oci:|dir:|file:, current source: %s", source)
	}

	return nil
}

func checkIsLiveMode() bool {
	dat, err := os.ReadFile(constants.CosLiveModeFile)
	if err != nil {
		slog.Debug("Error reading live_mode file", "error", err.Error())
		return false
	}

	return string(dat) == "1"
}

func setupLogger(ctx context.Context) log.Logger {
	return log.NewLogger(ctx)
}
