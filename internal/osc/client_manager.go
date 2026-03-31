package osc

import (
	"sync"

	"github.com/thomassaison/outscale-mcp/internal/config"
)

// ClientManager manages multiple Outscale API clients for different profiles.
type ClientManager struct {
	mu            sync.Mutex
	clients       map[string]*Client
	profileConfig *config.ProfileConfig
}

// NewClientManager creates a new client manager with the given profile configuration.
func NewClientManager(pc *config.ProfileConfig) *ClientManager {
	return &ClientManager{
		clients:       make(map[string]*Client),
		profileConfig: pc,
	}
}

// GetClient returns a client for the specified profile.
// If profile is empty, uses the default profile.
// Clients are cached for reuse.
func (cm *ClientManager) GetClient(profile string) (*Client, error) {
	if profile == "" {
		profile = cm.profileConfig.DefaultProfile
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.clients[profile]; ok {
		return client, nil
	}

	cfg, err := cm.profileConfig.GetProfile(profile)
	if err != nil {
		return nil, err
	}

	client, err := NewWithCredentials(cfg.AccessKey, cfg.SecretKey, cfg.Region)
	if err != nil {
		return nil, err
	}

	cm.clients[profile] = client
	return client, nil
}

// DefaultClient returns the client for the default profile.
func (cm *ClientManager) DefaultClient() (*Client, error) {
	return cm.GetClient("")
}

// ListProfiles returns the list of available profiles.
func (cm *ClientManager) ListProfiles() []string {
	return cm.profileConfig.ListProfiles()
}

// DefaultProfile returns the name of the default profile.
func (cm *ClientManager) DefaultProfile() string {
	return cm.profileConfig.DefaultProfile
}
