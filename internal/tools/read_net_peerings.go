package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadNetPeerings(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_net_peerings",
		mcp.WithDescription(`List and inspect Net Peerings in your Outscale account.

Use this tool to:
- Check Net peering configurations
- View peering states and connections
- Find peerings by ID, source Net, or accepter Net`),
		mcp.WithString("net_peering_ids",
			mcp.Description("Filter by Net Peering IDs (comma-separated)"),
		),
		mcp.WithString("source_net_ids",
			mcp.Description("Filter by source Net IDs (comma-separated)"),
		),
		mcp.WithString("accepter_net_ids",
			mcp.Description("Filter by accepter Net IDs (comma-separated)"),
		),
		mcp.WithString("state_names",
			mcp.Description("Filter by states: pending-acceptance, active, rejected, failed, expired, deleted (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadNetPeerings(authCtx, client, req, profile)
		})
	})
}

func handleReadNetPeerings(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersNetPeering{}
	args := req.Params.Arguments

	if netPeeringIds := getString(args, "net_peering_ids"); netPeeringIds != "" {
		filters.SetNetPeeringIds(parseCommaSeparated(netPeeringIds))
	}
	if sourceNetIds := getString(args, "source_net_ids"); sourceNetIds != "" {
		filters.SetSourceNetNetIds(parseCommaSeparated(sourceNetIds))
	}
	if accepterNetIds := getString(args, "accepter_net_ids"); accepterNetIds != "" {
		filters.SetAccepterNetNetIds(parseCommaSeparated(accepterNetIds))
	}
	if stateNames := getString(args, "state_names"); stateNames != "" {
		filters.SetStateNames(parseCommaSeparated(stateNames))
	}

	readReq := osc.ReadNetPeeringsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.NetPeeringApi.ReadNetPeerings(authCtx).ReadNetPeeringsRequest(readReq).Execute()
	if err != nil {
		return formatError("read net peerings", err), nil
	}

	netPeerings := make([]map[string]interface{}, 0)
	if read.NetPeerings != nil {
		for _, np := range *read.NetPeerings {
			netPeerings = append(netPeerings, formatNetPeering(np))
		}
	}

	response := map[string]interface{}{
		"net_peerings": netPeerings,
		"count":        len(netPeerings),
		"profile":      profile,
		"request_id":   safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatNetPeering(np osc.NetPeering) map[string]interface{} {
	result := map[string]interface{}{
		"net_peering_id": safeString(np.NetPeeringId),
	}

	if np.SourceNet != nil {
		result["source_net"] = map[string]interface{}{
			"net_id":     safeString(np.SourceNet.NetId),
			"account_id": safeString(np.SourceNet.AccountId),
			"ip_range":   safeString(np.SourceNet.IpRange),
		}
	}

	if np.AccepterNet != nil {
		result["accepter_net"] = map[string]interface{}{
			"net_id":     safeString(np.AccepterNet.NetId),
			"account_id": safeString(np.AccepterNet.AccountId),
			"ip_range":   safeString(np.AccepterNet.IpRange),
		}
	}

	if np.State != nil {
		result["state"] = map[string]interface{}{
			"name":    safeString(np.State.Name),
			"message": safeString(np.State.Message),
		}
	}

	if np.ExpirationDate.IsSet() {
		result["expiration_date"] = np.ExpirationDate.Get()
	}

	if np.Tags != nil {
		tags := make([]map[string]interface{}, 0)
		for _, tag := range *np.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		result["tags"] = tags
	}

	return result
}
