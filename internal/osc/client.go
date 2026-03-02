// Package osc provides a wrapper around the Outscale SDK client.
package osc

import (
	"context"

	osc "github.com/outscale/osc-sdk-go/v2"
)

// Client wraps the Outscale SDK API client.
type Client struct {
	API       *osc.APIClient
	configEnv *osc.ConfigEnv
}

// New creates a new Outscale API client from environment variables.
// Required environment variables:
//   - OSC_ACCESS_KEY: Your Outscale access key
//   - OSC_SECRET_KEY: Your Outscale secret key
//   - OSC_REGION: The region to use (e.g., "eu-west-2")
//
// Optional environment variables:
//   - OSC_ENDPOINT_API: Custom API endpoint
func New() (*Client, error) {
	configEnv := osc.NewConfigEnv()
	config, err := configEnv.Configuration()
	if err != nil {
		return nil, err
	}

	client := osc.NewAPIClient(config)

	return &Client{
		API:       client,
		configEnv: configEnv,
	}, nil
}

// NewWithCredentials creates a new Outscale API client with explicit credentials.
func NewWithCredentials(accessKey, secretKey, region string) (*Client, error) {
	configEnv := osc.NewConfigEnv()
	configEnv.AccessKey = &accessKey
	configEnv.SecretKey = &secretKey
	configEnv.Region = &region

	config, err := configEnv.Configuration()
	if err != nil {
		return nil, err
	}

	client := osc.NewAPIClient(config)

	return &Client{
		API:       client,
		configEnv: configEnv,
	}, nil
}

// Context returns a context with authentication configured.
func (c *Client) Context(ctx context.Context) (context.Context, error) {
	return c.configEnv.Context(ctx)
}
