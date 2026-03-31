package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterListProfiles(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_list_profiles",
		mcp.WithDescription(`List all available Outscale profiles.

Use this tool to:
- See which profiles are available for use
- Find the default profile name
- Switch between different Outscale accounts`),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListProfiles(ctx, clientManager)
	})
}

func handleListProfiles(ctx context.Context, clientManager *oscclient.ClientManager) (*mcp.CallToolResult, error) {
	profiles := clientManager.ListProfiles()
	defaultProfile := clientManager.DefaultProfile()

	result := make([]map[string]interface{}, 0, len(profiles))
	for _, name := range profiles {
		result = append(result, map[string]interface{}{
			"name":    name,
			"default": name == defaultProfile,
		})
	}

	response := map[string]interface{}{
		"profiles":        result,
		"count":           len(profiles),
		"default_profile": defaultProfile,
	}

	return formatResult(response)
}
