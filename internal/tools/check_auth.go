package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	osc "github.com/outscale/osc-sdk-go/v2"
	oscclient "github.com/thomassaison/outscale-mcp/internal/osc"
)

// RegisterCheckAuth registers the authentication verification tool.
func RegisterCheckAuth(s *server.MCPServer, clientManager *oscclient.ClientManager) {
	tool := mcp.NewTool("osc_check_auth",
		mcp.WithDescription("Verify Outscale API credentials and retrieve account information. Use this to test connectivity and authentication before other operations."),
		mcp.WithString("profile",
			mcp.Description("Profile name to use (optional, uses default if not specified)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return withClient(ctx, clientManager, req, handleCheckAuth)
	})
}

func handleCheckAuth(authCtx context.Context, client *oscclient.Client, profile string) (*mcp.CallToolResult, error) {
	read, _, err := client.API.AccountApi.ReadAccounts(authCtx).ReadAccountsRequest(osc.ReadAccountsRequest{}).Execute()
	if err != nil {
		return formatError("verify authentication", err), nil
	}

	if read.Accounts == nil || len(*read.Accounts) == 0 {
		return mcp.NewToolResultText("Authentication successful but no account information returned"), nil
	}

	accounts := *read.Accounts
	account := accounts[0]

	response := map[string]interface{}{
		"status":       "authenticated",
		"profile":      profile,
		"account_id":   safeString(account.AccountId),
		"customer_id":  safeString(account.CustomerId),
		"email":        safeString(account.Email),
		"company_name": safeString(account.CompanyName),
		"first_name":   safeString(account.FirstName),
		"last_name":    safeString(account.LastName),
		"request_id":   safeResponseId(read.ResponseContext),
	}

	return formatResult(response)
}
