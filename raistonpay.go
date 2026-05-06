// Package raistonpay provides a Go client for the RaistonPay agent payments platform.
package raistonpay

// Version is the current SDK version.
const Version = "0.1.0"

// Config holds the configuration for the RaistonPay client.
type Config struct {
	APIKey  string
	BaseURL string
}

// Client is the RaistonPay API client.
type Client struct {
	config Config
}

// NewClient creates a new RaistonPay client with the given configuration.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.raistonpay.com"
	}
	return &Client{config: cfg}
}
