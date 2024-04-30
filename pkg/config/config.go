package config

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/spf13/viper"
)

const defaultVersion = "v1.0"

type Config struct {
	// +optional, specify llmos os configurations
	OS LLMOS `json:"os,omitempty" yaml:"os,omitempty"`
	// +optional, specify llmos install configurations
	Install Install `json:"install,omitempty" yaml:"install,omitempty"`
	// +optional, specify llmos config file location (file path or http URL)
	ConfigURL string `json:"config-url,omitempty" yaml:"config-url,omitempty"`
	// +optional, specify the configuration version
	Version string `json:"version" yaml:"version"`
	// +optional, specify if the configuration is for debug purposes
	Debug bool `json:"debug,omitempty" yaml:"debug,omitempty"`
}

type LLMOS struct {
	SSHAuthorizedKeys    []string          `json:"ssh-authorized-keys,omitempty" yaml:"ssh-authorized-keys,omitempty"`
	WriteFiles           []File            `json:"write-files,omitempty" yaml:"write-files,omitempty"`
	Hostname             string            `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Modules              []string          `json:"modules,omitempty" yaml:"modules,omitempty"`
	Sysctl               map[string]string `json:"sysctl,omitempty" yaml:"sysctl,omitempty"`
	Username             string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string            `json:"password,omitempty" yaml:"password,omitempty"`
	NTPServers           []string          `json:"ntp-servers,omitempty" yaml:"ntp-servers,omitempty"`
	DNSNameservers       []string          `json:"dns-nameservers,omitempty" yaml:"dns-nameservers,omitempty"`
	Environment          map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	PersistentStatePaths []string          `json:"persistent-state-paths,omitempty" yaml:"persistent-state-paths,omitempty"`
	K3SConfig            `json:",inline,omitempty" yaml:",inline,omitempty"`
}

type K3SConfig struct {
	Token          string   `json:"token,omitempty" yaml:"token,omitempty"`
	NodeExternalIP string   `json:"node-external-ip,omitempty" yaml:"node-external-ip,omitempty"`
	NodeLabel      []string `json:"node-label,omitempty" yaml:"node-label,omitempty"`
}

type File struct {
	Encoding           string `json:"encoding" yaml:"encoding"`
	Content            string `json:"content" yaml:"content"`
	Owner              string `json:"owner" yaml:"owner"`
	Path               string `json:"path" yaml:"path"`
	RawFilePermissions string `json:"permissions" yaml:"permissions"`
}

type Install struct {
	Device     string   `json:"device" yaml:"device" binding:"required"`
	Silent     bool     `json:"silent,omitempty" yaml:"silent,omitempty"`
	ISOUrl     string   `json:"iso,omitempty" yaml:"iso,omitempty"`
	SystemURI  string   `json:"system-uri,omitempty" yaml:"system-uri,omitempty"`
	PowerOff   bool     `json:"poweroff,omitempty" yaml:"poweroff,omitempty"`
	Debug      bool     `json:"debug,omitempty" yaml:"debug,omitempty"`
	TTY        string   `json:"tty,omitempty" yaml:"tty,omitempty"`
	DataDevice string   `json:"data-device,omitempty" yaml:"data-device,omitempty"`
	Env        []string `json:"env,omitempty" yaml:"env,omitempty"`
	Reboot     bool     `json:"reboot,omitempty" yaml:"reboot,omitempty"`
	ConfigDir  string   `json:"config-dir,omitempty" yaml:"config-dir,omitempty"`
}

type Upgrade struct {
	Source          string `json:"source,omitempty" yaml:"source,omitempty"`
	UpgradeRecovery bool   `json:"upgrade-recovery,omitempty" yaml:"upgrade-recovery,omitempty"`
	HostDir         string `json:"host-dir,omitempty" yaml:"host-dir,omitempty"`
	Debug           bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
	Force           bool   `json:"force,omitempty" yaml:"force,omitempty"`
	Dev             bool   `json:"dev,omitempty" yaml:"dev,omitempty"`
}

func NewLLMOSConfig() *Config {
	debug := viper.GetBool("debug")
	return &Config{
		Version: defaultVersion,
		Debug:   debug,
		Install: Install{
			Debug: debug,
		},
	}
}

func (c *Config) DeepCopy() (*Config, error) {
	newConf := NewLLMOSConfig()
	if err := mergo.Merge(newConf, c, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *c, c, err.Error())
	}
	return newConf, nil
}

func (i *Install) DeepCopy() (*Install, error) {
	install := &Install{}
	if err := mergo.Merge(install, i, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *i, i, err.Error())
	}
	return install, nil
}

func (l *LLMOS) DeepCopy() (*LLMOS, error) {
	llmos := &LLMOS{}
	if err := mergo.Merge(llmos, l, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *l, l, err.Error())
	}
	return llmos, nil
}

func (c *Config) ToCosInstallEnv() ([]string, error) {
	return ToEnv("LLMOS_", c.Install)
}

func (c *Config) HasDataPartition() bool {
	return c.Install.DataDevice != ""
}

func (c *Config) GetK3sNodeLabels() []string {
	if c.OS.NodeLabel == nil {
		return []string{}
	}
	return c.OS.NodeLabel
}

func (c *Config) GetK3sDisabledComponents() []string {
	return []string{
		"cloud-controller",
	}
}

func (c *Config) GetK3sNodeExternalIP() string {
	return c.OS.NodeExternalIP
}

func (c *Config) Merge(cfg *Config) error {
	if err := mergo.Merge(c, cfg, mergo.WithAppendSlice); err != nil {
		return err
	}
	return nil
}
