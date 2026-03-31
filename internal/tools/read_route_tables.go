package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadRouteTables(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_route_tables",
		mcp.WithDescription(`List and inspect Route Tables in your Outscale account.

Use this tool to:
- Check route table configurations
- Inspect routes and their destinations
- View subnet associations
- Find route tables by ID, Net ID, or linked subnet`),
		mcp.WithString("route_table_ids",
			mcp.Description("Filter by Route Table IDs (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net/VPC IDs (comma-separated)"),
		),
		mcp.WithString("link_route_table_ids",
			mcp.Description("Filter by Link Route Table IDs (association IDs, comma-separated)"),
		),
		mcp.WithString("link_subnet_ids",
			mcp.Description("Filter by Subnet IDs linked to route tables (comma-separated)"),
		),
		mcp.WithString("route_gateway_ids",
			mcp.Description("Filter by gateway IDs specified in routes (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadRouteTables(authCtx, client, req, profile)
		})
	})
}

func handleReadRouteTables(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersRouteTable{}
	args := req.Params.Arguments

	if routeTableIds := getString(args, "route_table_ids"); routeTableIds != "" {
		filters.SetRouteTableIds(parseCommaSeparated(routeTableIds))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if linkRouteTableIds := getString(args, "link_route_table_ids"); linkRouteTableIds != "" {
		filters.SetLinkRouteTableLinkRouteTableIds(parseCommaSeparated(linkRouteTableIds))
	}
	if linkSubnetIds := getString(args, "link_subnet_ids"); linkSubnetIds != "" {
		filters.SetLinkSubnetIds(parseCommaSeparated(linkSubnetIds))
	}
	if gatewayIds := getString(args, "route_gateway_ids"); gatewayIds != "" {
		filters.SetRouteGatewayIds(parseCommaSeparated(gatewayIds))
	}

	readReq := osc.ReadRouteTablesRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.RouteTableApi.ReadRouteTables(authCtx).ReadRouteTablesRequest(readReq).Execute()
	if err != nil {
		return formatError("read route tables", err), nil
	}

	routeTables := make([]map[string]interface{}, 0)
	if read.RouteTables != nil {
		for _, rt := range *read.RouteTables {
			routeTables = append(routeTables, formatRouteTable(rt))
		}
	}

	response := map[string]interface{}{
		"route_tables": routeTables,
		"count":        len(routeTables),
		"profile":      profile,
		"request_id":   safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatRouteTable(rt osc.RouteTable) map[string]interface{} {
	result := map[string]interface{}{
		"route_table_id": safeString(rt.RouteTableId),
		"net_id":         safeString(rt.NetId),
	}

	if rt.LinkRouteTables != nil {
		links := make([]map[string]interface{}, 0)
		for _, link := range *rt.LinkRouteTables {
			links = append(links, map[string]interface{}{
				"link_route_table_id": safeString(link.LinkRouteTableId),
				"main":                safeBool(link.Main),
				"subnet_id":           safeString(link.SubnetId),
			})
		}
		result["link_route_tables"] = links
	}

	if rt.Routes != nil {
		routes := make([]map[string]interface{}, 0)
		for _, route := range *rt.Routes {
			routes = append(routes, map[string]interface{}{
				"destination_ip_range": safeString(route.DestinationIpRange),
				"gateway_id":           safeString(route.GatewayId),
				"nat_service_id":       safeString(route.NatServiceId),
				"net_peering_id":       safeString(route.NetPeeringId),
				"vm_id":                safeString(route.VmId),
				"state":                safeString(route.State),
				"creation_method":      safeString(route.CreationMethod),
			})
		}
		result["routes"] = routes
	}

	if rt.Tags != nil {
		tags := make([]map[string]interface{}, 0)
		for _, tag := range *rt.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		result["tags"] = tags
	}

	return result
}
