package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadSubnets registers the Subnet inspection tool.
func RegisterReadSubnets(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_subnets",
		mcp.WithDescription(`List and inspect Subnets in your Outscale account.

Use this tool to:
- Check subnet configurations
- Debug subnet availability issues
- Inspect CIDR blocks and availability zones
- Find subnets by ID, Net ID, or state`),
		mcp.WithString("subnet_ids",
			mcp.Description("Filter by Subnet IDs (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net/VPC IDs (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, available (comma-separated)"),
		),
		mcp.WithString("availability_zones",
			mcp.Description("Filter by availability zones (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadSubnets(ctx, client, req)
	})
}

func handleReadSubnets(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersSubnet{}
	args := req.Params.Arguments

	if subnetIds := getString(args, "subnet_ids"); subnetIds != "" {
		filters.SetSubnetIds(parseCommaSeparated(subnetIds))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetStates(parseCommaSeparated(states))
	}
	if azs := getString(args, "availability_zones"); azs != "" {
		filters.SetSubregionNames(parseCommaSeparated(azs))
	}

	readReq := osc.ReadSubnetsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.SubnetApi.ReadSubnets(authCtx).ReadSubnetsRequest(readReq).Execute()
	if err != nil {
		return formatError("read subnets", err), nil
	}

	subnets := make([]map[string]interface{}, 0)
	if read.Subnets != nil {
		for _, subnet := range *read.Subnets {
			subnets = append(subnets, map[string]interface{}{
				"subnet_id":      safeString(subnet.SubnetId),
				"state":          safeString(subnet.State),
				"ip_range":       safeString(subnet.IpRange),
				"net_id":         safeString(subnet.NetId),
				"subregion_name": safeString(subnet.SubregionName),
				"available_ips":  safeInt(subnet.AvailableIpsCount),
			})
		}
	}

	response := map[string]interface{}{
		"subnets":    subnets,
		"count":      len(subnets),
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
