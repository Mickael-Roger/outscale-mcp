package tools

import (
	"context"
	"encoding/base64"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

func RegisterReadConsoleOutput(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_read_console_output",
		mcp.WithDescription(`Get the console output (boot logs) of a virtual machine.

Use this tool to:
- Debug VM boot issues
- Check cloud-init logs
- Diagnose startup errors
- View kernel messages`),
		mcp.WithString("vm_id",
			mcp.Description("The ID of the VM (required)"),
			mcp.Required(),
		),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, func(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
			return handleReadConsoleOutput(authCtx, client, req, profile)
		})
	})
}

func handleReadConsoleOutput(authCtx context.Context, client *oscclient.Client, req mcp.CallToolRequest, profile string) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	vmId := getString(args, "vm_id")
	if vmId == "" {
		return mcp.NewToolResultText("Error: vm_id parameter is required"), nil
	}

	readReq := osc.NewReadConsoleOutputRequest(vmId)

	read, _, err := client.API.VmApi.ReadConsoleOutput(authCtx).ReadConsoleOutputRequest(*readReq).Execute()
	if err != nil {
		return formatError("read console output", err), nil
	}

	consoleOutput := ""
	if read.ConsoleOutput != nil {
		decoded, err := base64.StdEncoding.DecodeString(*read.ConsoleOutput)
		if err != nil {
			return mcp.NewToolResultText("Failed to decode console output: " + err.Error()), nil
		}
		consoleOutput = string(decoded)
	}

	response := map[string]interface{}{
		"vm_id":          safeString(read.VmId),
		"console_output": consoleOutput,
		"profile":        profile,
		"request_id":     safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
