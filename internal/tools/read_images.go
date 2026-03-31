package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterReadImages registers the Image inspection tool.
func RegisterReadImages(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_images",
		mcp.WithDescription(`List and inspect machine images (OMIs) in your Outscale account.

Use this tool to:
- Check image states and availability
- Find suitable images for VM creation
- Debug image-related issues
- Inspect image configurations`),
		mcp.WithString("image_ids",
			mcp.Description("Filter by Image IDs (comma-separated)"),
		),
		mcp.WithString("image_names",
			mcp.Description("Filter by Image names (comma-separated)"),
		),
		mcp.WithString("states",
			mcp.Description("Filter by states: pending, available, failed (comma-separated)"),
		),
		mcp.WithString("architectures",
			mcp.Description("Filter by architectures: i386, x86_64 (comma-separated)"),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadImages(authCtx, client, req, profile)
		})
	})
}

func handleReadImages(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	filters := osc.FiltersImage{}
	args := req.Params.Arguments

	if imageIds := getString(args, "image_ids"); imageIds != "" {
		filters.SetImageIds(parseCommaSeparated(imageIds))
	}
	if imageNames := getString(args, "image_names"); imageNames != "" {
		filters.SetImageNames(parseCommaSeparated(imageNames))
	}
	if states := getString(args, "states"); states != "" {
		filters.SetStates(parseCommaSeparated(states))
	}
	if archs := getString(args, "architectures"); archs != "" {
		filters.SetArchitectures(parseCommaSeparated(archs))
	}

	readReq := osc.ReadImagesRequest{}
	readReq.SetFilters(filters)

	read, _, err := client.API.ImageApi.ReadImages(authCtx).ReadImagesRequest(readReq).Execute()
	if err != nil {
		return formatError("read images", err), nil
	}

	images := make([]map[string]interface{}, 0)
	if read.Images != nil {
		for _, img := range *read.Images {
			images = append(images, map[string]interface{}{
				"image_id":         safeString(img.ImageId),
				"name":             safeString(img.ImageName),
				"description":      safeString(img.Description),
				"state":            safeString(img.State),
				"architecture":     safeString(img.Architecture),
				"root_device_type": safeString(img.RootDeviceType),
				"creation_date":    safeString(img.CreationDate),
				"account_id":       safeString(img.AccountId),
			})
		}
	}

	response := map[string]interface{}{
		"images":     images,
		"count":      len(images),
		"profile":    profile,
		"request_id": safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
