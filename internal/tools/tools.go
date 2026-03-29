// Package tools provides MCP tools for debugging Outscale resources.
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterAll registers all debugging tools with the MCP server.
func RegisterAll(s *server.MCPServer, client *oscclient.Client) {
	RegisterCheckAuth(s, client)
	RegisterReadVMs(s, client)
	RegisterReadVolumes(s, client)
	RegisterReadNets(s, client)
	RegisterReadSubnets(s, client)
	RegisterReadRouteTables(s, client)
	RegisterReadSecurityGroups(s, client)
	RegisterReadPublicIps(s, client)
	RegisterReadApiLogs(s, client)
	RegisterReadQuotas(s, client)
	RegisterReadImages(s, client)
	RegisterReadVmState(s, client)
	RegisterReadInternetServices(s, client)
	RegisterReadNatServices(s, client)
	RegisterReadNetPeerings(s, client)
	RegisterReadNetAccessPoints(s, client)
	RegisterReadLoadBalancers(s, client)
}

// formatResult formats a result as JSON text.
func formatResult(data interface{}) (*mcp.CallToolResult, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("failed to format result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(bytes)), nil
}

// formatError creates an error result with context.
func formatError(operation string, err error) *mcp.CallToolResult {
	return mcp.NewToolResultText(fmt.Sprintf("failed to %s: %v", operation, err))
}

// parseCommaSeparated splits a comma-separated string into a slice.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	for i := 0; i < len(s); i++ {
		j := i
		for j < len(s) && s[j] != ',' {
			j++
		}
		part := trimSpace(s[i:j])
		if part != "" {
			result = append(result, part)
		}
		i = j
	}
	return result
}

// trimSpace trims whitespace from a string.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && isWhitespace(s[start]) {
		start++
	}
	for end > start && isWhitespace(s[end-1]) {
		end--
	}
	return s[start:end]
}

// isWhitespace checks if a byte is whitespace.
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// getString extracts a string parameter from the request arguments.
func getString(args map[string]interface{}, key string) string {
	if val, ok := args[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
