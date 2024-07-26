package retry

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func Retry(ctx context.Context, interval time.Duration, args []string) error {
	for {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			logrus.Errorf("will retry failed command %v: %v", args, err)
			select {
			case <-time.After(interval):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
}
