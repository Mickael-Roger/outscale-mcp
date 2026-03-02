package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadApiLogs registers the API logs inspection tool.
func RegisterReadApiLogs(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_api_logs",
		mcp.WithDescription(`Query API access logs for debugging and auditing.

Use this tool to:
- Debug API call failures
- Track API usage patterns
- Investigate access issues
- Find requests by API name, date, or HTTP code`),
		mcp.WithString("call_names",
			mcp.Description("Filter by API call names (comma-separated, e.g., ReadVms, CreateVms)"),
		),
		mcp.WithString("date_after",
			mcp.Description("Filter by date after (ISO 8601 format, e.g., 2024-01-15T00:00:00.000Z)"),
		),
		mcp.WithString("date_before",
			mcp.Description("Filter by date before (ISO 8601 format, e.g., 2024-01-15T23:59:59.000Z)"),
		),
		mcp.WithString("response_status_codes",
			mcp.Description("Filter by HTTP status codes (comma-separated, e.g., 200, 400, 500)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadApiLogs(ctx, client, req)
	})
}

func handleReadApiLogs(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersApiLog{}
	args := req.Params.Arguments

	if callNames := getString(args, "call_names"); callNames != "" {
		filters.SetQueryCallNames(parseCommaSeparated(callNames))
	}
	if dateAfter := getString(args, "date_after"); dateAfter != "" {
		filters.SetQueryDateAfter(dateAfter)
	}
	if dateBefore := getString(args, "date_before"); dateBefore != "" {
		filters.SetQueryDateBefore(dateBefore)
	}
	if statusCodes := getString(args, "response_status_codes"); statusCodes != "" {
		codes := parseCommaSeparated(statusCodes)
		intCodes := make([]int32, len(codes))
		for i, code := range codes {
			var val int32
			for _, c := range code {
				if c >= '0' && c <= '9' {
					val = val*10 + (c - '0')
				}
			}
			intCodes[i] = val
		}
		filters.SetResponseStatusCodes(intCodes)
	}

	readReq := osc.ReadApiLogsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.ApiLogApi.ReadApiLogs(authCtx).ReadApiLogsRequest(readReq).Execute()
	if err != nil {
		return formatError("read API logs", err), nil
	}

	logs := make([]map[string]interface{}, 0)
	if read.Logs != nil {
		for _, log := range *read.Logs {
			logs = append(logs, map[string]interface{}{
				"call_name":     safeString(log.QueryCallName),
				"date":          safeString(log.QueryDate),
				"response_code": safeInt32(log.ResponseStatusCode),
				"ip_address":    safeString(log.QueryIpAddress),
				"access_key":    safeString(log.QueryAccessKey),
				"request_id":    safeString(log.RequestId),
				"user_agent":    safeString(log.QueryUserAgent),
				"response_size": safeInt(log.ResponseSize),
			})
		}
	}

	response := map[string]interface{}{
		"logs":       logs,
		"count":      len(logs),
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
