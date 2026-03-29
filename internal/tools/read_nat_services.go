package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadNatServices(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_nat_services",
		mcp.WithDescription(`List and inspect NAT Services in your Outscale account.

Use this tool to:
- Check NAT service configurations
- View public IPs associated with NAT services
- Find NAT services by ID, Net, or Subnet`),
		mcp.WithString("nat_service_ids",
			mcp.Description("Filter by NAT Service IDs (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net IDs (comma-separated)"),
		),
		mcp.WithString("subnet_ids",
			mcp.Description("Filter by Subnet IDs (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, available, deleting, deleted (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadNatServices(ctx, client, req)
	})
}

func handleReadNatServices(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersNatService{}
	args := req.Params.Arguments

	if natServiceIds := getString(args, "nat_service_ids"); natServiceIds != "" {
		filters.SetNatServiceIds(parseCommaSeparated(natServiceIds))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if subnetIds := getString(args, "subnet_ids"); subnetIds != "" {
		filters.SetSubnetIds(parseCommaSeparated(subnetIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetStates(parseCommaSeparated(states))
	}

	readReq := osc.ReadNatServicesRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.NatServiceApi.ReadNatServices(authCtx).ReadNatServicesRequest(readReq).Execute()
	if err != nil {
		return formatError("read nat services", err), nil
	}

	natServices := make([]map[string]interface{}, 0)
	if read.NatServices != nil {
		for _, ns := range *read.NatServices {
			natServices = append(natServices, formatNatService(ns))
		}
	}

	response := map[string]interface{}{
		"nat_services": natServices,
		"count":        len(natServices),
		"request_id":   safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatNatService(ns osc.NatService) map[string]interface{} {
	result := map[string]interface{}{
		"nat_service_id": safeString(ns.NatServiceId),
		"net_id":         safeString(ns.NetId),
		"subnet_id":      safeString(ns.SubnetId),
		"state":          safeString(ns.State),
	}

	if ns.PublicIps != nil {
		publicIps := make([]map[string]interface{}, 0)
		for _, ip := range *ns.PublicIps {
			publicIps = append(publicIps, map[string]interface{}{
				"public_ip_id": safeString(ip.PublicIpId),
				"public_ip":    safeString(ip.PublicIp),
			})
		}
		result["public_ips"] = publicIps
	}

	if ns.Tags != nil {
		tags := make([]map[string]interface{}, 0)
		for _, tag := range *ns.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		result["tags"] = tags
	}

	return result
}
