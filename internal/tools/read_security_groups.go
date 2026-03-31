package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadSecurityGroups registers the Security Group inspection tool.
func RegisterReadSecurityGroups(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_security_groups",
		mcp.WithDescription(`List and inspect Security Groups in your Outscale account.

Use this tool to:
- Check security group rules (inbound/outbound)
- Debug connectivity and firewall issues
- Inspect port and protocol configurations
- Find security groups by ID, name, or Net ID`),
		mcp.WithString("security_group_ids",
			mcp.Description("Filter by Security Group IDs (comma-separated)"),
		),
		mcp.WithString("security_group_names",
			mcp.Description("Filter by Security Group names (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net/VPC IDs (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadSecurityGroups(authCtx, client, req, profile)
		})
	})
}

func handleReadSecurityGroups(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersSecurityGroup{}
	args := req.Params.Arguments

	if sgIds := getString(args, "security_group_ids"); sgIds != "" {
		filters.SetSecurityGroupIds(parseCommaSeparated(sgIds))
	}
	if sgNames := getString(args, "security_group_names"); sgNames != "" {
		filters.SetSecurityGroupNames(parseCommaSeparated(sgNames))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}

	readReq := osc.ReadSecurityGroupsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.SecurityGroupApi.ReadSecurityGroups(authCtx).ReadSecurityGroupsRequest(readReq).Execute()
	if err != nil {
		return formatError("read security groups", err), nil
	}

	sgs := make([]map[string]interface{}, 0)
	if read.SecurityGroups != nil {
		for _, sg := range *read.SecurityGroups {
			sgs = append(sgs, map[string]interface{}{
				"security_group_id":   safeString(sg.SecurityGroupId),
				"security_group_name": safeString(sg.SecurityGroupName),
				"description":         safeString(sg.Description),
				"net_id":              safeString(sg.NetId),
				"inbound_rules":       formatSecurityGroupRules(sg.InboundRules),
				"outbound_rules":      formatSecurityGroupRules(sg.OutboundRules),
			})
		}
	}

	response := map[string]interface{}{
		"security_groups": sgs,
		"count":           len(sgs),
		"profile":         profile,
		"request_id":      safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatSecurityGroupRules(rules *[]osc.SecurityGroupRule) []map[string]interface{} {
	if rules == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*rules))
	for i, rule := range *rules {
		ipRanges := []string{}
		if rule.IpRanges != nil {
			for _, ip := range *rule.IpRanges {
				ipRanges = append(ipRanges, ip)
			}
		}

		result[i] = map[string]interface{}{
			"protocol":  safeString(rule.IpProtocol),
			"from_port": safeInt(rule.FromPortRange),
			"to_port":   safeInt(rule.ToPortRange),
			"ip_ranges": ipRanges,
		}
	}
	return result
}
