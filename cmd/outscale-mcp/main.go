// Package main provides the entry point for the Outscale MCP server.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/thomassaison/outscale-mcp/internal/config"
	"github.com/thomassaison/outscale-mcp/internal/osc"
	"github.com/thomassaison/outscale-mcp/internal/tools"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create Outscale API client
	client, err := osc.NewWithCredentials(cfg.AccessKey, cfg.SecretKey, cfg.Region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Create context with auth
	ctx := context.Background()

	// Create MCP server
	s := server.NewMCPServer(
		"Outscale Debug Tools",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools
	tools.RegisterAll(s, client)

	// Start the server
	fmt.Println("Starting Outscale MCP server...")
	fmt.Printf("Region: %s\n", cfg.Region)

	// Verify auth on startup
	_, err = client.Context(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
