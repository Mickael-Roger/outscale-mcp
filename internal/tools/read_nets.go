package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadNets registers the Net/VPC inspection tool.
func RegisterReadNets(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_nets",
		mcp.WithDescription(`List and inspect Nets (VPCs) in your Outscale account.

Use this tool to:
- Check VPC/Net configurations
- Debug network connectivity issues
- Inspect CIDR blocks and DHCP options
- Find Nets by ID or state`),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net/VPC IDs (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, available (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadNets(authCtx, client, req, profile)
		})
	})
}

func handleReadNets(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersNet{}
	args := req.Params.Arguments

	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetStates(parseCommaSeparated(states))
	}

	readReq := osc.ReadNetsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.NetApi.ReadNets(authCtx).ReadNetsRequest(readReq).Execute()
	if err != nil {
		return formatError("read Nets", err), nil
	}

	nets := make([]map[string]interface{}, 0)
	if read.Nets != nil {
		for _, net := range *read.Nets {
			nets = append(nets, map[string]interface{}{
				"net_id":          safeString(net.NetId),
				"state":           safeString(net.State),
				"ip_range":        safeString(net.IpRange),
				"dhcp_options_id": safeString(net.DhcpOptionsSetId),
			})
		}
	}

	response := map[string]interface{}{
		"nets":       nets,
		"count":      len(nets),
		"profile":    profile,
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
