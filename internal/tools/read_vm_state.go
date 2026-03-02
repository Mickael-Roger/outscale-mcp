package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadVmState registers the VM state inspection tool.
func RegisterReadVmState(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_vm_state",
		mcp.WithDescription(`Get detailed state information for specific VMs.

Use this tool to:
- Check VM health and status
- Debug VM state transitions
- Monitor VM lifecycle events
- Get maintenance event information`),
		mcp.WithString("vm_ids",
			mcp.Required(),
			mcp.Description("VM IDs to check (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadVmState(ctx, client, req)
	})
}

func handleReadVmState(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	args := req.Params.Arguments
	vmIds := getString(args, "vm_ids")
	if vmIds == "" {
		return mcp.NewToolResultText("vm_ids parameter is required"), nil
	}

	filters := osc.FiltersVmsState{}
	filters.SetVmIds(parseCommaSeparated(vmIds))

	readReq := osc.ReadVmsStateRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.VmApi.ReadVmsState(authCtx).ReadVmsStateRequest(readReq).Execute()
	if err != nil {
		return formatError("read VM state", err), nil
	}

	vmStates := make([]map[string]interface{}, 0)
	if read.VmStates != nil {
		for _, state := range *read.VmStates {
			vmStates = append(vmStates, map[string]interface{}{
				"vm_id":              safeString(state.VmId),
				"state":              safeString(state.VmState),
				"subregion_name":     safeString(state.SubregionName),
				"maintenance_events": extractMaintenanceEvents(state.MaintenanceEvents),
			})
		}
	}

	response := map[string]interface{}{
		"vm_states":  vmStates,
		"count":      len(vmStates),
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func extractMaintenanceEvents(events *[]osc.MaintenanceEvent) []map[string]interface{} {
	if events == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*events))
	for i, event := range *events {
		result[i] = map[string]interface{}{
			"code":        safeString(event.Code),
			"description": safeString(event.Description),
			"not_before":  safeString(event.NotBefore),
			"not_after":   safeString(event.NotAfter),
		}
	}
	return result
}
