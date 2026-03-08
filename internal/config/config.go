package config

import (
	"fmt"
	"os"
)

type Config struct {
	KeaURL   string
	KeaToken string
	DNSURL   string
	DNSToken string
	SubnetID string
}

func Load() (*Config, error) {
	c := &Config{
		KeaURL:   os.Getenv("AWKTO_KEA_URL"),
		KeaToken: os.Getenv("AWKTO_KEA_TOKEN"),
		DNSURL:   os.Getenv("AWKTO_DNS_URL"),
		DNSToken: os.Getenv("AWKTO_DNS_TOKEN"),
		SubnetID: os.Getenv("AWKTO_SUBNET_ID"),
	}
	if c.SubnetID == "" {
		c.SubnetID = "1"
	}
	return c, nil
}

func (c *Config) RequireKea() error {
	if c.KeaURL == "" {
		return fmt.Errorf("AWKTO_KEA_URL is not set")
	}
	return nil
}

func (c *Config) RequireDNS() error {
	if c.DNSURL == "" {
		return fmt.Errorf("AWKTO_DNS_URL is not set")
	}
	return nil
}
