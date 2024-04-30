package system

import (
	"path/filepath"
)

const (
	HostRootDir = "/host"
	// LocalDir represents where llmos persistent installation is located
	LocalDir = "/var/lib/llmos"
	// ConfigDir represents where persistent configuration is located
	ConfigDir = "/etc/llmos"
	// ConfigFileDir represents where llmos configuration is located
	ConfigFileDir = ConfigDir + "/config.d"
	// ExtraDataDir represents where llmos extra data disk path is located
	ExtraDataDir = "/var/lib/llmos-data"
	// StateDir represents where cos ephemeral state is located
	StateDir = "/run/elemental"
)

var (
	hostRoot     = HostRootDir
	localDir     = LocalDir
	configDir    = ConfigDir
	extraDataDir = ExtraDataDir
	stateDir     = StateDir
)

func HostRootPath(elem ...string) string {
	return filepath.Join(hostRoot, filepath.Join(elem...))
}

func LocalPath(elem ...string) string {
	return filepath.Join(localDir, filepath.Join(elem...))
}

func ConfigPath(elem ...string) string {
	return filepath.Join(configDir, filepath.Join(elem...))
}

func ExtraDataPath(elem ...string) string {
	return filepath.Join(extraDataDir, filepath.Join(elem...))
}

func StatePath(elem ...string) string {
	return filepath.Join(stateDir, filepath.Join(elem...))
}
