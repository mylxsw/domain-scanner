package main

import (
	"fmt"
	"os"
	"strings"
)

// NamecheapConfig holds the configuration for Namecheap API
type NamecheapConfig struct {
	ApiUser  string
	ApiKey   string
	Username string
	ClientIp string
	ApiBase  string
}

// FromEnv loads configuration from environment variables
func FromEnv() (*NamecheapConfig, error) {
	apiUser := strings.TrimSpace(os.Getenv("NAMECHEAP_API_USER"))
	apiKey := strings.TrimSpace(os.Getenv("NAMECHEAP_API_KEY"))
	username := strings.TrimSpace(os.Getenv("NAMECHEAP_USERNAME"))
	clientIp := strings.TrimSpace(os.Getenv("NAMECHEAP_CLIENT_IP"))
	apiBase := strings.TrimSpace(os.Getenv("NAMECHEAP_API_BASE"))

	if username == "" {
		username = apiUser
	}
	if apiBase == "" {
		apiBase = "https://api.namecheap.com/xml.response"
	}

	var missing []string
	if apiUser == "" {
		missing = append(missing, "NAMECHEAP_API_USER")
	}
	if apiKey == "" {
		missing = append(missing, "NAMECHEAP_API_KEY")
	}
	if clientIp == "" {
		missing = append(missing, "NAMECHEAP_CLIENT_IP")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s. Please refer to README.md for configuration", strings.Join(missing, ", "))
	}

	return &NamecheapConfig{
		ApiUser:  apiUser,
		ApiKey:   apiKey,
		Username: username,
		ClientIp: clientIp,
		ApiBase:  apiBase,
	}, nil
}
