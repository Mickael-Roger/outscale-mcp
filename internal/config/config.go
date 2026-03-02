// Package config provides configuration management for the Outscale MCP server.
package config

import (
	"errors"
	"os"
)

// Config holds the Outscale API credentials and settings.
type Config struct {
	AccessKey string
	SecretKey string
	Region    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		AccessKey: os.Getenv("OSC_ACCESS_KEY"),
		SecretKey: os.Getenv("OSC_SECRET_KEY"),
		Region:    os.Getenv("OSC_REGION"),
	}

	if cfg.Region == "" {
		cfg.Region = "eu-west-2"
	}

	if cfg.AccessKey == "" {
		return nil, errors.New("OSC_ACCESS_KEY environment variable is required")
	}
	if cfg.SecretKey == "" {
		return nil, errors.New("OSC_SECRET_KEY environment variable is required")
	}

	return cfg, nil
}
