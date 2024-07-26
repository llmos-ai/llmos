//go:build windows
// +build windows

package applyinator

import (
	"os"

	"github.com/sirupsen/logrus"
)

// reconcileFilePermissions abstracts out the file permissions checks and are a no op on Windows
func reconcileFilePermissions(path string, uid int, gid int, perm os.FileMode) error {
	logrus.Debugf("windows file permissions for %s will not be reconciled to %d:%d %d", path, uid, gid, perm)
	return nil
}
