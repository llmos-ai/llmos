package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	yipSchema "github.com/mudler/yip/pkg/schema"
	"gopkg.in/yaml.v1"
)

const (
	ntpdService         = "systemd-timesyncd"
	timeWaitSyncService = "systemd-time-wait-sync"

	K3sConfigFile   = "/etc/rancher/k3s/config.yaml"
	K3sManifestPath = "/var/lib/rancher/k3s/server/manifests/"
)

// ConvertToCos converts LLMOSConfig into the cOS configuration
func ConvertToCos(cfg *LLMOSConfig) (*yipSchema.YipConfig, error) {
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

	// k3s related configs
	if err = addInitK3sStage(cfg, &initramfs); err != nil {
		return nil, err
	}

	if err = addLLMOSManifests(cfg, &initramfs); err != nil {
		return nil, err
	}

	// OS configs
	username := cfg.OS.Username
	initramfs.Users[username] = yipSchema.User{
		PasswordHash: cfg.OS.Password,
		Groups:       []string{"root"},
		Homedir:      fmt.Sprintf("/home/%s", username),
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
			RootfsStage.String():    {rootfs},
			InitramfsStage.String(): {initramfs},
			NetworkStage.String():   {afterNetwork},
		},
	}, nil
}

func overwriteRootfsStage(cfg *LLMOSConfig, stage *yipSchema.Stage) error {
	content, err := Render("cos_rootfs.yaml", cfg)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(content), stage); err != nil {
		return err
	}

	return nil
}

func addInitK3sStage(cfg *LLMOSConfig, stage *yipSchema.Stage) error {
	manifestConfig, err := Render("k3s-config.yaml", cfg)
	if err != nil {
		return err
	}
	stage.Files = append(stage.Files,
		yipSchema.File{
			Path:        K3sConfigFile,
			Content:     manifestConfig,
			Permissions: 0644,
			Owner:       0,
			Group:       0,
		},
	)
	return nil
}

func addLLMOSManifests(cfg *LLMOSConfig, stage *yipSchema.Stage) error {
	for _, templateName := range []string{
		"llmos-namespace.yaml",
		"ollama-service.yaml",
		"open-webui.yaml",
	} {
		fileContent, err := Render(templateName, cfg)
		if err != nil {
			return err
		}

		stage.Files = append(stage.Files,
			yipSchema.File{
				Path:        filepath.Join(K3sManifestPath, templateName),
				Content:     fileContent,
				Permissions: 0600,
				Owner:       0,
				Group:       0,
			},
		)
	}
	return nil
}
