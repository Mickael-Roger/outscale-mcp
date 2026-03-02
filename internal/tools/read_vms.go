package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadVMs registers the VM inspection tool.
func RegisterReadVMs(s *server.MCPServer, client *oscclient.Client) {
	tool := mcp.NewTool("osc_read_vms",
		mcp.WithDescription(`List and inspect virtual machines (VMs) in your Outscale account.

Use this tool to:
- Check VM states (pending, running, stopping, stopped, terminated)
- Debug VM connectivity issues
- Inspect VM configurations (type, image, network)
- Find VMs by ID, name, or state`),
		mcp.WithString("vm_ids",
			mcp.Description("Filter by VM IDs (comma-separated)"),
		),
		mcp.WithString("image_ids",
			mcp.Description("Filter by Image IDs (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, running, stopping, stopped, terminated (comma-separated)"),
		),
		mcp.WithString("net_ids",
			mcp.Description("Filter by VPC/Net IDs (comma-separated)"),
		),
		mcp.WithString("keypair_names",
			mcp.Description("Filter by keypair names (comma-separated)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleReadVMs(ctx, client, req)
	})
}

func handleReadVMs(ctx context.Context, client *oscclient.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authCtx, err := client.Context(ctx)
	if err != nil {
		return mcp.NewToolResultText("Authentication failed: " + err.Error()), nil
	}

	filters := osc.FiltersVm{}
	args := req.Params.Arguments

	if vmIds := getString(args, "vm_ids"); vmIds != "" {
		filters.SetVmIds(parseCommaSeparated(vmIds))
	}
	if imageIds := getString(args, "image_ids"); imageIds != "" {
		filters.SetImageIds(parseCommaSeparated(imageIds))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetVmStateNames(parseCommaSeparated(states))
	}
	if netIds := getString(args, "net_ids"); netIds != "" {
		filters.SetNetIds(parseCommaSeparated(netIds))
	}
	if keypairNames := getString(args, "keypair_names"); keypairNames != "" {
		filters.SetKeypairNames(parseCommaSeparated(keypairNames))
	}

	readReq := osc.ReadVmsRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.VmApi.ReadVms(authCtx).ReadVmsRequest(readReq).Execute()
	if err != nil {
		return formatError("read VMs", err), nil
	}

	vms := make([]map[string]interface{}, 0)
	if read.Vms != nil {
		for _, vm := range *read.Vms {
			vms = append(vms, map[string]interface{}{
				"vm_id":           safeString(vm.VmId),
				"name":            extractTagName(vm.Tags),
				"state":           safeString(vm.State),
				"vm_type":         safeString(vm.VmType),
				"image_id":        safeString(vm.ImageId),
				"net_id":          safeString(vm.NetId),
				"subnet_id":       safeString(vm.SubnetId),
				"public_ip":       safeString(vm.PublicIp),
				"private_ip":      safeString(vm.PrivateIp),
				"security_groups": extractSecurityGroupIds(vm.SecurityGroups),
				"keypair":         safeString(vm.KeypairName),
				"creation_date":   safeString(vm.CreationDate),
			})
		}
	}

	response := map[string]interface{}{
		"vms":        vms,
		"count":      len(vms),
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}

func extractTagName(tags *[]osc.ResourceTag) string {
	if tags == nil {
		return ""
	}
	for _, tag := range *tags {
		if tag.Key == "Name" {
			return tag.Value
		}
	}
	return ""
}

func extractSecurityGroupIds(sgs *[]osc.SecurityGroupLight) []string {
	if sgs == nil {
		return []string{}
	}
	result := make([]string, len(*sgs))
	for i, sg := range *sgs {
		result[i] = safeString(sg.SecurityGroupId)
	}
	return result
}
