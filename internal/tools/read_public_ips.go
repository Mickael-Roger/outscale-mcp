package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadPublicIps registers the Public IP inspection tool.
func RegisterReadPublicIps(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_public_ips",
		mcp.WithDescription(`List and inspect Public IP addresses in your Outscale account.

Use this tool to:
- Check public IP associations
- Debug connectivity issues related to public IPs
- Find which VM a public IP is linked to
- Check for unassociated public IPs`),
		mcp.WithString("public_ip_ids",
			mcp.Description("Filter by Public IP IDs (comma-separated)"),
		),
		mcp.WithString("public_ips",
			mcp.Description("Filter by Public IP addresses (comma-separated)"),
		),
		mcp.WithString("linked_vm_ids",
			mcp.Description("Filter by VM IDs that IPs are linked to (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadPublicIps(authCtx, client, req, profile)
		})
	})
}

func handleReadPublicIps(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersPublicIp{}
	args := req.Params.Arguments

	if ipIds := getString(args, "public_ip_ids"); ipIds != "" {
		filters.SetPublicIpIds(parseCommaSeparated(ipIds))
	}
	if ips := getString(args, "public_ips"); ips != "" {
		filters.SetPublicIps(parseCommaSeparated(ips))
	}
	if vmIds := getString(args, "linked_vm_ids"); vmIds != "" {
		filters.SetVmIds(parseCommaSeparated(vmIds))
	}

	readReq := osc.ReadPublicIpsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.PublicIpApi.ReadPublicIps(authCtx).ReadPublicIpsRequest(readReq).Execute()
	if err != nil {
		return formatError("read public IPs", err), nil
	}

	publicIps := make([]map[string]interface{}, 0)
	if read.PublicIps != nil {
		for _, ip := range *read.PublicIps {
			publicIps = append(publicIps, map[string]interface{}{
				"public_ip_id":         safeString(ip.PublicIpId),
				"public_ip":            safeString(ip.PublicIp),
				"linked_vm_id":         safeString(ip.VmId),
				"linked_vm_private_ip": safeString(ip.PrivateIp),
			})
		}
	}

	response := map[string]interface{}{
		"public_ips": publicIps,
		"count":      len(publicIps),
		"profile":    profile,
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
