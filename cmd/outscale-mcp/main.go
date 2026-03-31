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
	pc, err := config.LoadProfileConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create client manager
	clientManager := osc.NewClientManager(pc)

	// Create context with auth
	ctx := context.Background()

	// Create MCP server
	s := server.NewMCPServer(
		"Outscale Debug Tools",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools
	tools.RegisterAll(s, clientManager)

	// Start the server
	fmt.Println("Starting Outscale MCP server...")
	fmt.Printf("Profiles available: %v\n", clientManager.ListProfiles())
	fmt.Printf("Default profile: %s\n", clientManager.DefaultProfile())

	// Verify auth on startup
	client, err := clientManager.DefaultClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get default client: %v\n", err)
		os.Exit(1)
	}
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
