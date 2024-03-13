package config

import (
	"fmt"

	"github.com/imdario/mergo"
)

type LLMOSConfig struct {
	OS      LLMOS   `json:"os,omitempty"`
	Install Install `json:"install,omitempty"`
	LLM     LLM     `json:"llm,omitempty"`
}

type LLM struct {
	model        string `json:"model,omitempty"`
	startAtLogin string `json:"startAtLogin,omitempty"`
	HTTPSPort    int    `json:"httpsPort,omitempty"`
}

type LLMOS struct {
	SSHAuthorizedKeys []string `json:"sshAuthorizedKeys,omitempty"`
	WriteFiles        []File   `json:"writeFiles,omitempty"`
	Hostname          string   `json:"hostname,omitempty"`
	Runcmd            []string `json:"runCmd,omitempty"`
	Bootcmd           []string `json:"bootCmd,omitempty"`
	Initcmd           []string `json:"initCmd,omitempty"`
	Env               []string `yaml:"env,omitempty"`

	Username             string            `json:"username,omitempty"`
	Password             string            `json:"password,omitempty"`
	Modules              []string          `json:"modules,omitempty"`
	Sysctls              map[string]string `json:"sysctls,omitempty"`
	NTPServers           []string          `json:"ntpServers,omitempty"`
	DNSNameservers       []string          `json:"dnsNameservers,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
	PersistentStatePaths []string          `json:"persistentStatePaths,omitempty"`
}

type File struct {
	Encoding           string `json:"encoding"`
	Content            string `json:"content"`
	Owner              string `json:"owner"`
	Path               string `json:"path"`
	RawFilePermissions string `json:"permissions"`
}

type Install struct {
	//ForceEFI  bool   `json:"forceEfi,omitempty"`
	Device     string   `json:"device,omitempty"`
	ConfigURL  string   `json:"configUrl,omitempty"`
	Silent     bool     `json:"silent,omitempty"`
	ISOURL     string   `json:"isoUrl,omitempty"`
	PowerOff   bool     `json:"powerOff,omitempty"`
	NoFormat   bool     `json:"noFormat,omitempty"`
	Debug      bool     `json:"debug,omitempty"`
	TTY        string   `json:"tty,omitempty"`
	DataDevice string   `json:"dataDevice,omitempty"`
	Env        []string `json:"env,omitempty"`
	Reboot     bool     `json:"reboot,omitempty"`
}

func NewLLMOSConfig() *LLMOSConfig {
	return &LLMOSConfig{}
}

func (c *LLMOSConfig) DeepCopy() (*LLMOSConfig, error) {
	newConf := NewLLMOSConfig()
	if err := mergo.Merge(newConf, c, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *c, c, err.Error())
	}
	return newConf, nil
}

func (c *LLMOSConfig) ToCosInstallEnv() ([]string, error) {
	return ToEnv("LLMOS_", c.Install)
}

type Stage string

const (
	RootfsStage    Stage = "rootfs"
	InitramfsStage Stage = "initramfs"
	NetworkStage   Stage = "network"
)

func (n Stage) String() string {
	return string(n)
}
