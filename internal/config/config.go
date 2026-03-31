// Package config provides configuration management for the Outscale MCP server.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const DefaultRegion = "eu-west-2"

// Config holds the Outscale API credentials and settings.
type Config struct {
	AccessKey string
	SecretKey string
	Region    string
}

// Profile represents a single Outscale profile in the config file.
type Profile struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

// ConfigFile represents the structure of .osc/config.json
type ConfigFile map[string]Profile

// ProfileConfig holds the configuration for profile-based authentication.
type ProfileConfig struct {
	DefaultProfile string
	Profiles       ConfigFile
	ConfigFilePath string
}

// LoadProfileConfig loads configuration supporting multiple profiles.
// Priority:
// 1. If OSC_ACCESS_KEY, OSC_SECRET_KEY, OSC_REGION are set -> use as default
// 2. Otherwise, load from OSC_CONFIG_FILE or ~/.osc/config.json
func LoadProfileConfig() (*ProfileConfig, error) {
	pc := &ProfileConfig{
		Profiles: make(ConfigFile),
	}

	accessKey := os.Getenv("OSC_ACCESS_KEY")
	secretKey := os.Getenv("OSC_SECRET_KEY")
	region := os.Getenv("OSC_REGION")

	if region == "" {
		region = DefaultRegion
	}

	if accessKey != "" && secretKey != "" {
		pc.DefaultProfile = "default"
		pc.Profiles["default"] = Profile{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    region,
		}
		return pc, nil
	}

	configPath := os.Getenv("OSC_CONFIG_FILE")
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.New("cannot determine home directory")
		}
		configPath = filepath.Join(home, ".osc", "config.json")
	}

	pc.ConfigFilePath = configPath

	info, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %q - create it or set OSC_ACCESS_KEY/OSC_SECRET_KEY", configPath)
	}

	if info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("config file has insecure permissions; run: chmod 600 %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.New("cannot read config file")
	}

	if err := json.Unmarshal(data, &pc.Profiles); err != nil {
		return nil, errors.New("cannot parse config file")
	}

	if len(pc.Profiles) == 0 {
		return nil, errors.New("no profiles found in config file")
	}

	if _, ok := pc.Profiles["default"]; ok {
		pc.DefaultProfile = "default"
	} else {
		names := pc.ListProfiles()
		pc.DefaultProfile = names[0]
	}

	return pc, nil
}

// GetProfile returns the configuration for a specific profile.
func (pc *ProfileConfig) GetProfile(name string) (*Config, error) {
	if name == "" {
		name = pc.DefaultProfile
	}

	profile, ok := pc.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile not found: %q (available: %v)", name, pc.ListProfiles())
	}

	region := profile.Region
	if region == "" {
		region = DefaultRegion
	}

	return &Config{
		AccessKey: profile.AccessKey,
		SecretKey: profile.SecretKey,
		Region:    region,
	}, nil
}

// ListProfiles returns the list of available profile names in sorted order.
func (pc *ProfileConfig) ListProfiles() []string {
	names := make([]string, 0, len(pc.Profiles))
	for name := range pc.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
