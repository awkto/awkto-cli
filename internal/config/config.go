package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the resolved runtime configuration used by commands.
type Config struct {
	KeaURL   string
	KeaToken string
	DNSURL   string
	DNSToken string
	SubnetID string
}

// Server represents a single server entry in the config file.
type Server struct {
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	Token    string `yaml:"token,omitempty"`
	SubnetID string `yaml:"subnet_id,omitempty"`
}

// ConfigFile represents the on-disk config file structure.
type ConfigFile struct {
	Defaults map[string]string    `yaml:"defaults,omitempty"`
	Servers  map[string]Server    `yaml:"servers,omitempty"`
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
		return &ConfigFile{
			Defaults: make(map[string]string),
			Servers:  make(map[string]Server),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{
				Defaults: make(map[string]string),
				Servers:  make(map[string]Server),
			}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cf ConfigFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	if cf.Defaults == nil {
		cf.Defaults = make(map[string]string)
	}
	if cf.Servers == nil {
		cf.Servers = make(map[string]Server)
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

// Load returns the resolved Config using default servers from the config file,
// with env vars as highest-priority overrides.
func Load() (*Config, error) {
	return loadForType("", "")
}

// LoadForDNS loads configuration for a DNS command.
// If serverName is non-empty, it loads that specific server and validates it is type "dns".
// Otherwise it uses the default dns server from the config file.
// Env vars always override.
func LoadForDNS(serverName string) (*Config, error) {
	return loadForType(serverName, "dns")
}

// LoadForKea loads configuration for a Kea command.
// If serverName is non-empty, it loads that specific server and validates it is type "kea".
// Otherwise it uses the default kea server from the config file.
// Env vars always override.
func LoadForKea(serverName string) (*Config, error) {
	return loadForType(serverName, "kea")
}

func loadForType(serverName string, requiredType string) (*Config, error) {
	c := &Config{}

	cf, err := LoadConfigFile()
	if err != nil {
		return nil, err
	}

	if serverName != "" {
		// Load a specific named server
		srv, ok := cf.Servers[serverName]
		if !ok {
			return nil, fmt.Errorf("server %q not found", serverName)
		}
		if requiredType != "" && srv.Type != requiredType {
			return nil, fmt.Errorf("server %q is type %s, but this command requires %s", serverName, srv.Type, requiredType)
		}
		applyServer(c, srv)
	} else {
		// Load defaults for the required type(s)
		if requiredType == "" || requiredType == "dns" {
			if defaultName, ok := cf.Defaults["dns"]; ok && defaultName != "" {
				if srv, ok := cf.Servers[defaultName]; ok {
					c.DNSURL = srv.URL
					c.DNSToken = srv.Token
				}
			}
		}
		if requiredType == "" || requiredType == "kea" {
			if defaultName, ok := cf.Defaults["kea"]; ok && defaultName != "" {
				if srv, ok := cf.Servers[defaultName]; ok {
					c.KeaURL = srv.URL
					c.KeaToken = srv.Token
					if srv.SubnetID != "" {
						c.SubnetID = srv.SubnetID
					}
				}
			}
		}
	}

	if c.SubnetID == "" {
		c.SubnetID = "1"
	}

	return c, nil
}

func applyServer(c *Config, srv Server) {
	switch srv.Type {
	case "dns":
		c.DNSURL = srv.URL
		c.DNSToken = srv.Token
	case "kea":
		c.KeaURL = srv.URL
		c.KeaToken = srv.Token
		if srv.SubnetID != "" {
			c.SubnetID = srv.SubnetID
		}
	}
}

func (c *Config) RequireKea() error {
	if c.KeaURL == "" {
		return fmt.Errorf("no kea server configured (configure with 'awkto server add')")
	}
	return nil
}

func (c *Config) RequireDNS() error {
	if c.DNSURL == "" {
		return fmt.Errorf("no dns server configured (configure with 'awkto server add')")
	}
	return nil
}
