package install

import (
	"log/slog"
	"strconv"
	"strings"

	yipSchema "github.com/mudler/yip/pkg/schema"
	"gopkg.in/yaml.v1"

	"github.com/oneblock-ai/llmos/pkg/config"
)

const (
	ntpdService         = "systemd-timesyncd"
	timeWaitSyncService = "systemd-time-wait-sync"
)

// ConvertToCos converts LLMOSConfig into the cOS configuration
func ConvertToCos(cfg *config.LLMOSConfig) (*yipSchema.YipConfig, error) {
	cfg, err := cfg.DeepCopy()
	if err != nil {
		return nil, err
	}

	// Overwrite rootfs layout
	rootfs := yipSchema.Stage{}
	if err = overwriteRootfsStage(cfg, &rootfs); err != nil {
		return nil, err
	}

	initramfs := yipSchema.Stage{
		Users:     make(map[string]yipSchema.User),
		TimeSyncd: make(map[string]string),
	}

	afterNetwork := yipSchema.Stage{
		SSHKeys: make(map[string][]string),
	}

	username := cfg.OS.Username
	initramfs.Users[username] = yipSchema.User{
		PasswordHash: cfg.OS.Password,
	}

	// Use modprobe to load modules as a workaround solution
	for _, module := range cfg.OS.Modules {
		initramfs.Commands = append(initramfs.Commands, "modprobe "+module)
	}

	initramfs.Sysctl = cfg.OS.Sysctls
	initramfs.Environment = cfg.OS.Environment

	// append write_files
	for _, ff := range cfg.OS.WriteFiles {
		perm, err := strconv.ParseUint(ff.RawFilePermissions, 8, 32)
		if err != nil {
			slog.Error("fail to parse permission, use default permission 600", err)
			perm = 0600
		}
		initramfs.Files = append(initramfs.Files, yipSchema.File{
			Path:        ff.Path,
			Content:     ff.Content,
			Encoding:    ff.Encoding,
			Permissions: uint32(perm),
			OwnerString: ff.Owner,
		})
	}

	// config hostname
	if len(cfg.OS.Hostname) > 0 {
		initramfs.Hostname = cfg.OS.Hostname
	}

	// set ntp servers
	if len(cfg.OS.NTPServers) > 0 {
		initramfs.TimeSyncd["NTP"] = strings.Join(cfg.OS.NTPServers, " ")
		initramfs.Systemctl.Enable = append(initramfs.Systemctl.Enable, ntpdService)
		initramfs.Systemctl.Enable = append(initramfs.Systemctl.Enable, timeWaitSyncService)
	}

	// set DNS nameservers
	if len(cfg.OS.DNSNameservers) > 0 {
		initramfs.Dns.Nameservers = cfg.OS.DNSNameservers
	}

	// set ssh authorized keys
	if len(cfg.OS.SSHAuthorizedKeys) > 0 {
		afterNetwork.SSHKeys[username] = cfg.OS.SSHAuthorizedKeys
	}

	return &yipSchema.YipConfig{
		Name: "LLMOS Installer Configuration",
		Stages: map[string][]yipSchema.Stage{
			config.RootfsStage.String():    {rootfs},
			config.InitramfsStage.String(): {initramfs},
			config.NetworkStage.String():   {afterNetwork},
		},
	}, nil
}

func overwriteRootfsStage(cfg *config.LLMOSConfig, stage *yipSchema.Stage) error {
	content, err := config.Render("cos_rootfs.yaml", cfg)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(content), stage); err != nil {
		return err
	}

	return nil
}
