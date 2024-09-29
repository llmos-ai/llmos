package applyinator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
)

const defaultDirectoryPermissions os.FileMode = 0755
const defaultFilePermissions os.FileMode = 0600

func writeBase64ContentToFile(file File) error {
	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return err
	}
	var fileMode os.FileMode
	if file.Permissions == "" {
		logrus.Debugf("[Applyinator] Requested file permission for %s was %s, defaulting to %d",
			file.Path, file.Permissions, defaultFilePermissions)
		fileMode = defaultFilePermissions
	} else {
		parsedPerm, err := parsePerm(file.Permissions)
		if err != nil {
			return err
		}
		fileMode = parsedPerm
	}
	return writeContentToFile(file.Path, file.UID, file.GID, fileMode, content)
}

func writeContentToFile(path string, uid int, gid int, perm os.FileMode, content []byte) error {
	if path == "" {
		return fmt.Errorf("path was empty")
	}

	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, content) {
		logrus.Debugf("[Applyinator] File %s does not need to be written", path)
	} else {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, defaultDirectoryPermissions); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, perm); err != nil {
			return err
		}
	}
	return reconcileFilePermissions(path, uid, gid, perm)
}

func createDirectory(file File) error {
	if !file.Directory {
		return fmt.Errorf("%s was not a directory", file.Path)
	}
	var fileMode os.FileMode
	if file.Permissions == "" {
		logrus.Debugf("[Applyinator] Requested file permission for %s was %s, defaulting to %d",
			file.Path, file.Permissions, defaultDirectoryPermissions)
		fileMode = defaultDirectoryPermissions
	} else {
		parsedPerm, err := parsePerm(file.Permissions)
		if err != nil {
			return err
		}
		fileMode = parsedPerm
	}

	if err := os.MkdirAll(file.Path, fileMode); err != nil {
		return err
	}

	return reconcileFilePermissions(file.Path, file.UID, file.GID, fileMode)
}

func parsePerm(perm string) (os.FileMode, error) {
	parsedPerm, err := strconv.ParseInt(perm, 8, 32)
	if err != nil {
		return defaultFilePermissions, err
	}
	return os.FileMode(parsedPerm), nil
}
