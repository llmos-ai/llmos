package system

import (
	"path/filepath"
)

const (
	HostRootDir = "/host"
	// DataDir represents where llmos persistent installation is located
	DataDir = "/var/lib/llmos"
	// ConfigDir represents where persistent configuration is located
	ConfigDir = "/etc/llmos"
	// DefaultConfigFile represents where the default llmos configuration is located
	DefaultConfigFile = ConfigDir + "/config.yaml"
	// ConfigFileDir represents where llmos configuration is located
	ConfigFileDir = ConfigDir + "/config.d"
	// ExtraDataDir represents where llmos extra data disk path is located
	ExtraDataDir = "/var/lib/llmos-data"
	// CosStateDir represents where cos ephemeral state is located
	CosStateDir = "/run/elemental"
)

var (
	hostRoot     = HostRootDir
	dataDir      = DataDir
	configDir    = ConfigDir
	extraDataDir = ExtraDataDir
	stateDir     = CosStateDir
)

func HostRootPath(elem ...string) string {
	return filepath.Join(hostRoot, filepath.Join(elem...))
}

func DataPath(elem ...string) string {
	return filepath.Join(dataDir, filepath.Join(elem...))
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
