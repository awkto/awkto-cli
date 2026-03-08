package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/awkto/awkto-cli/internal/config"
)

type KeaClient struct {
	url   string
	token string
}

type Reservation struct {
	IPAddress string `json:"ip-address"`
	HWAddress string `json:"hw-address"`
	Hostname  string `json:"hostname"`
	SubnetID  int    `json:"subnet_id"`
}

type ReservationCreate struct {
	IPAddress string `json:"ip_address"`
	HWAddress string `json:"hw_address"`
	Hostname  string `json:"hostname"`
	SubnetID  int    `json:"subnet_id"`
}

type Lease struct {
	IPAddress  string `json:"ip-address"`
	HWAddress  string `json:"hw-address"`
	Hostname   string `json:"hostname"`
	State      int    `json:"state"`
	SubnetID   int    `json:"subnet-id"`
	ValidLft   int    `json:"valid-lft"`
	Cltt       int    `json:"cltt"`
	FqdnFwd    bool   `json:"fqdn-fwd"`
	FqdnRev    bool   `json:"fqdn-rev"`
}

func NewKeaClient(cfg *config.Config) *KeaClient {
	return &KeaClient{
		url:   baseURL(cfg.KeaURL),
		token: cfg.KeaToken,
	}
}

// Reservations

func (c *KeaClient) ListReservations(subnetID string) ([]Reservation, error) {
	url := fmt.Sprintf("%s/api/reservations?subnet_id=%s", c.url, subnetID)
	body, status, err := doRequest(http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", status, string(body))
	}

	var result struct {
		Reservations []Reservation `json:"reservations"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result.Reservations, nil
}

func (c *KeaClient) CreateReservation(r ReservationCreate) error {
	url := fmt.Sprintf("%s/api/reservations", c.url)
	body, status, err := doRequest(http.MethodPost, url, c.token, r)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

func (c *KeaClient) DeleteReservation(ipAddress string) error {
	url := fmt.Sprintf("%s/api/reservation/%s", c.url, ipAddress)
	body, status, err := doRequest(http.MethodDelete, url, c.token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

// Leases

func (c *KeaClient) ListLeases(subnetID string) ([]Lease, error) {
	url := fmt.Sprintf("%s/api/leases?subnet_id=%s", c.url, subnetID)
	body, status, err := doRequest(http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", status, string(body))
	}

	var result struct {
		Leases []Lease `json:"leases"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result.Leases, nil
}

func (c *KeaClient) DeleteLeaseByIP(ipAddress string) error {
	url := fmt.Sprintf("%s/api/leases/ip/%s", c.url, ipAddress)
	body, status, err := doRequest(http.MethodDelete, url, c.token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}

func (c *KeaClient) DeleteLeaseByMAC(macAddress string) error {
	url := fmt.Sprintf("%s/api/leases/mac/%s", c.url, macAddress)
	body, status, err := doRequest(http.MethodDelete, url, c.token, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", status, string(body))
	}
	return nil
}
