package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/awkto/awkto-cli/internal/config"
)

type DNSClient struct {
	url   string
	token string
}

type DNSRecord struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	TTL    int      `json:"ttl"`
	FQDN   string   `json:"fqdn,omitempty"`
	Values []string `json:"values"`
}

type DNSRecordCreate struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	TTL    int      `json:"ttl"`
	Values []string `json:"values"`
}

type DNSRecordUpdate struct {
	TTL    int      `json:"ttl,omitempty"`
	Values []string `json:"values,omitempty"`
}

func NewDNSClient(cfg *config.Config) *DNSClient {
	return &DNSClient{
		url:   baseURL(cfg.DNSURL),
		token: cfg.DNSToken,
	}
}

func (c *DNSClient) ListRecords() ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/api/records", c.url)
	body, status, err := doRequest(http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", status, string(body))
	}

	var result struct {
		Records []DNSRecord `json:"records"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result.Records, nil
}

func (c *DNSClient) CreateRecord(r DNSRecordCreate) error {
	url := fmt.Sprintf("%s/api/records", c.url)
	body, status, err := doRequest(http.MethodPost, url, c.token, r)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

func (c *DNSClient) UpdateRecord(recordType, name string, update DNSRecordUpdate) error {
	url := fmt.Sprintf("%s/api/records/%s/%s", c.url, recordType, name)
	body, status, err := doRequest(http.MethodPut, url, c.token, update)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

func (c *DNSClient) DeleteRecord(recordType, name string) error {
	url := fmt.Sprintf("%s/api/records/%s/%s", c.url, recordType, name)
	body, status, err := doRequest(http.MethodDelete, url, c.token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

func (c *DNSClient) Health() error {
	url := fmt.Sprintf("%s/api/health", c.url)
	body, status, err := doRequest(http.MethodGet, url, c.token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}
