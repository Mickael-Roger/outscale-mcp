package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadNetAccessPoints(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_net_access_points",
		mcp.WithDescription(`List and inspect Net Access Points (VPC Endpoints) in your Outscale account.

Use this tool to:
- Check Net access point configurations
- View endpoint services and route table associations
- Find endpoints by ID, Net, or state`),
		mcp.WithString("net_access_point_ids",
			mcp.Description("Filter by Net Access Point IDs (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net IDs (comma-separated)"),
		),
		mcp.WithString("service_names",
			mcp.Description("Filter by service names (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, available, deleting, deleted (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadNetAccessPoints(authCtx, client, req, profile)
		})
	})
}

func handleReadNetAccessPoints(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersNetAccessPoint{}
	args := req.Params.Arguments

	if netAccessPointIds := getString(args, "net_access_point_ids"); netAccessPointIds != "" {
		filters.SetNetAccessPointIds(parseCommaSeparated(netAccessPointIds))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if serviceNames := getString(args, "service_names"); serviceNames != "" {
		filters.SetServiceNames(parseCommaSeparated(serviceNames))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetStates(parseCommaSeparated(states))
	}

	readReq := osc.ReadNetAccessPointsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.NetAccessPointApi.ReadNetAccessPoints(authCtx).ReadNetAccessPointsRequest(readReq).Execute()
	if err != nil {
		return formatError("read net access points", err), nil
	}

	netAccessPoints := make([]map[string]interface{}, 0)
	if read.NetAccessPoints != nil {
		for _, nap := range *read.NetAccessPoints {
			netAccessPoints = append(netAccessPoints, formatNetAccessPoint(nap))
		}
	}

	response := map[string]interface{}{
		"net_access_points": netAccessPoints,
		"count":             len(netAccessPoints),
		"profile":           profile,
		"request_id":        safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatNetAccessPoint(nap osc.NetAccessPoint) map[string]interface{} {
	result := map[string]interface{}{
		"net_access_point_id": safeString(nap.NetAccessPointId),
		"net_id":              safeString(nap.NetId),
		"service_name":        safeString(nap.ServiceName),
		"state":               safeString(nap.State),
	}

	if nap.RouteTableIds != nil {
		result["route_table_ids"] = *nap.RouteTableIds
	}

	if nap.Tags != nil {
		tags := make([]map[string]interface{}, 0)
		for _, tag := range *nap.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		result["tags"] = tags
	}

	return result
}
