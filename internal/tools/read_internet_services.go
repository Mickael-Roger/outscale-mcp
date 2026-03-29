package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadInternetServices(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_internet_services",
		mcp.WithDescription(`List and inspect Internet Services in your Outscale account.

Use this tool to:
- Check internet service configurations
- View Net attachments
- Find internet services by ID or attached Net`),
		mcp.WithString("internet_service_ids",
			mcp.Description("Filter by Internet Service IDs (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by Net IDs the internet services are attached to (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by attachment states (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadInternetServices(ctx, client, req)
	})
}

func handleReadInternetServices(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersInternetService{}
	args := req.Params.Arguments

	if internetServiceIds := getString(args, "internet_service_ids"); internetServiceIds != "" {
		filters.SetInternetServiceIds(parseCommaSeparated(internetServiceIds))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetLinkNetIds(parseCommaSeparated(netIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetLinkStates(parseCommaSeparated(states))
	}

	readReq := osc.ReadInternetServicesRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.InternetServiceApi.ReadInternetServices(authCtx).ReadInternetServicesRequest(readReq).Execute()
	if err != nil {
		return formatError("read internet services", err), nil
	}

	internetServices := make([]map[string]interface{}, 0)
	if read.InternetServices != nil {
		for _, is := range *read.InternetServices {
			internetServices = append(internetServices, formatInternetService(is))
		}
	}

	response := map[string]interface{}{
		"internet_services": internetServices,
		"count":             len(internetServices),
		"request_id":        safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func formatInternetService(is osc.InternetService) map[string]interface{} {
	result := map[string]interface{}{
		"internet_service_id": safeString(is.InternetServiceId),
		"net_id":              safeString(is.NetId),
		"state":               safeString(is.State),
	}

	if is.Tags != nil {
		tags := make([]map[string]interface{}, 0)
		for _, tag := range *is.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		result["tags"] = tags
	}

	return result
}
