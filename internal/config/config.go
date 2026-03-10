package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	KeaURL   string
	KeaToken string
	DNSURL   string
	DNSToken string
	SubnetID string
}

type Context struct {
	DNSURL   string `yaml:"dns_url,omitempty"`
	DNSToken string `yaml:"dns_token,omitempty"`
	KeaURL   string `yaml:"kea_url,omitempty"`
	KeaToken string `yaml:"kea_token,omitempty"`
	SubnetID string `yaml:"subnet_id,omitempty"`
}

type ConfigFile struct {
	CurrentContext string             `yaml:"current-context"`
	Contexts       map[string]Context `yaml:"contexts"`
}

func ConfigFilePath() string {
	if p := os.Getenv("AWKTO_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".awkto", "config.yaml")
}

func LoadConfigFile() (*ConfigFile, error) {
	path := ConfigFilePath()
	if path == "" {
		return &ConfigFile{Contexts: make(map[string]Context)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{Contexts: make(map[string]Context)}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cf ConfigFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	if cf.Contexts == nil {
		cf.Contexts = make(map[string]Context)
	}
	return &cf, nil
}

func SaveConfigFile(cf *ConfigFile) error {
	path := ConfigFilePath()
	if path == "" {
		return fmt.Errorf("cannot determine config file path")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(cf)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func Load() (*Config, error) {
	c := &Config{}

	// First, try loading from config file
	cf, err := LoadConfigFile()
	if err != nil {
		return nil, err
	}

	if cf.CurrentContext != "" {
		ctx, ok := cf.Contexts[cf.CurrentContext]
		if ok {
			c.DNSURL = ctx.DNSURL
			c.DNSToken = ctx.DNSToken
			c.KeaURL = ctx.KeaURL
			c.KeaToken = ctx.KeaToken
			c.SubnetID = ctx.SubnetID
		}
	}

	// Then, let env vars override
	if v := os.Getenv("AWKTO_DNS_URL"); v != "" {
		c.DNSURL = v
	}
	if v := os.Getenv("AWKTO_DNS_TOKEN"); v != "" {
		c.DNSToken = v
	}
	if v := os.Getenv("AWKTO_KEA_URL"); v != "" {
		c.KeaURL = v
	}
	if v := os.Getenv("AWKTO_KEA_TOKEN"); v != "" {
		c.KeaToken = v
	}
	if v := os.Getenv("AWKTO_SUBNET_ID"); v != "" {
		c.SubnetID = v
	}

	if c.SubnetID == "" {
		c.SubnetID = "1"
	}

	return c, nil
}

func (c *Config) RequireKea() error {
	if c.KeaURL == "" {
		return fmt.Errorf("AWKTO_KEA_URL is not set (set via env var or config context)")
	}
	return nil
}

func (c *Config) RequireDNS() error {
	if c.DNSURL == "" {
		return fmt.Errorf("AWKTO_DNS_URL is not set (set via env var or config context)")
	}
	return nil
}
