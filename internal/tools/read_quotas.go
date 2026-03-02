package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadQuotas registers the Quota inspection tool.
func RegisterReadQuotas(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_quotas",
		mcp.WithDescription(`List and inspect account quotas in your Outscale account.

Use this tool to:
- Check resource limits
- Debug quota exceeded errors
- Plan capacity requirements
- Verify quota usage`),
		mcp.WithString("quota_names",
			mcp.Description("Filter by quota names (comma-separated)"),
		),
		mcp.WithString("quota_types",
			mcp.Description("Filter by quota types (comma-separated)"),
		),
		mcp.WithString("collections",
			mcp.Description("Filter by collection/group names (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadQuotas(ctx, client, req)
	})
}

func handleReadQuotas(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersQuota{}
	args := req.Params.Arguments

	if quotaNames := getString(args, "quota_names"); quotaNames != "" {
		filters.SetQuotaNames(parseCommaSeparated(quotaNames))
	}
	if quotaTypes := getString(args, "quota_types"); quotaTypes != "" {
		filters.SetQuotaTypes(parseCommaSeparated(quotaTypes))
	}
	if collections := getString(args, "collections"); collections != "" {
		filters.SetCollections(parseCommaSeparated(collections))
	}

	readReq := osc.ReadQuotasRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.QuotaApi.ReadQuotas(authCtx).ReadQuotasRequest(readReq).Execute()
	if err != nil {
		return formatError("read quotas", err), nil
	}

	quotas := make([]map[string]interface{}, 0)
	if read.QuotaTypes != nil {
		for _, qt := range *read.QuotaTypes {
			if qt.Quotas != nil {
				for _, q := range *qt.Quotas {
					quotas = append(quotas, map[string]interface{}{
						"quota_name":        safeString(q.Name),
						"quota_type":        safeString(qt.QuotaType),
						"short_description": safeString(q.ShortDescription),
						"max_value":         safeInt(q.MaxValue),
						"used_value":        safeInt(q.UsedValue),
					})
				}
			}
		}
	}

	response := map[string]interface{}{
		"quotas":     quotas,
		"count":      len(quotas),
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
