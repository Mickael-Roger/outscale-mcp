package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadVolumes registers the volume inspection tool.
func RegisterReadVolumes(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_volumes",
		mcp.WithDescription(`List and inspect block storage volumes in your Outscale account.

Use this tool to:
- Check volume states (creating, available, in-use, deleting, deleted, error)
- Debug volume attachment issues
- Inspect volume configurations (size, type, IOPS)
- Find volumes linked to specific VMs`),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
		mcp.WithString("volume_ids",
			mcp.Description("Filter by volume IDs (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: creating, available, in-use, deleting, error (comma-separated)"),
		),
		mcp.WithString("vm_ids",
			mcp.Description("Filter by VM IDs that volumes are attached to (comma-separated)"),
		),
		mcp.WithString("volume_types",
			mcp.Description("Filter by volume types: standard, gp2, io1 (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadVolumes(authCtx, client, req, profile)
		})
	})
}

func handleReadVolumes(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersVolume{}
	args := req.Params.Arguments

	if volumeIds := getString(args, "volume_ids"); volumeIds != "" {
		filters.SetVolumeIds(parseCommaSeparated(volumeIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetVolumeStates(parseCommaSeparated(states))
	}
	if vmIds := getString(args, "vm_ids"); vmIds != "" {
		filters.SetLinkVolumeVmIds(parseCommaSeparated(vmIds))
	}
	if types := getString(args, "volume_types"); types != "" {
		filters.SetVolumeTypes(parseCommaSeparated(types))
	}

	readReq := osc.ReadVolumesRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.VolumeApi.ReadVolumes(authCtx).ReadVolumesRequest(readReq).Execute()
	if err != nil {
		return formatError("read volumes", err), nil
	}

	volumes := make([]map[string]interface{}, 0)
	if read.Volumes != nil {
		for _, vol := range *read.Volumes {
			volumes = append(volumes, map[string]interface{}{
				"volume_id":     safeString(vol.VolumeId),
				"state":         safeString(vol.State),
				"size_gb":       safeInt(vol.Size),
				"volume_type":   safeString(vol.VolumeType),
				"iops":          safeInt(vol.Iops),
				"snapshot_id":   safeString(vol.SnapshotId),
				"subregion":     safeString(vol.SubregionName),
				"linked_to":     extractLinkedVolumes(vol.LinkedVolumes),
				"creation_date": safeString(vol.CreationDate),
			})
		}
	}

	response := map[string]interface{}{
		"volumes":    volumes,
		"count":      len(volumes),
		"profile":    profile,
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func extractLinkedVolumes(linked *[]osc.LinkedVolume) []map[string]interface{} {
	if linked == nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*linked))
	for i, lv := range *linked {
		result[i] = map[string]interface{}{
			"vm_id":  safeString(lv.VmId),
			"device": safeString(lv.DeviceName),
			"state":  safeString(lv.State),
		}
	}
	return result
}
