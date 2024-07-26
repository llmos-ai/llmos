//go:build !windows
// +build !windows

package applyinator

import (
	"os"

	"github.com/sirupsen/logrus"
)

// reconcileFilePermissions abstracts out the file permissions checks that only works on Linux.
func reconcileFilePermissions(path string, uid int, gid int, perm os.FileMode) error {
	logrus.Debugf("[Applyinator] Reconciling file permissions for %s to %d:%d %d", path, uid, gid, perm)
	if err := os.Chmod(path, perm); err != nil {
		return err
	}
	return os.Chown(path, uid, gid)
}
