# Outscale MCP Server

A Model Context Protocol (MCP) server for debugging Outscale API resources. This server provides read-only tools to inspect and debug your Outscale cloud infrastructure.

## Prerequisites

- Go 1.21 or later
- Outscale account with API credentials

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd outscale-mcp

# Build
go build -o outscale-mcp ./cmd/outscale-mcp
```

## Configuration

Set the following environment variables with your Outscale credentials:

```bash
export OSC_ACCESS_KEY="your-access-key"
export OSC_SECRET_KEY="your-secret-key"
export OSC_REGION="eu-west-2"  # Optional, defaults to eu-west-2
```

## Usage

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "outscale": {
      "command": "/path/to/outscale-mcp",
      "env": {
        "OSC_ACCESS_KEY": "your-access-key",
        "OSC_SECRET_KEY": "your-secret-key",
        "OSC_REGION": "eu-west-2"
      }
    }
  }
}
```

### With Other MCP Clients

Run the server directly:

```bash
./outscale-mcp
```

The server communicates via stdio using the MCP protocol.

## Available Tools

| Tool | Description |
|------|-------------|
| `osc_check_auth` | Verify API credentials are valid |
| `osc_read_vms` | List and inspect virtual machines |
| `osc_read_vm_state` | Get detailed state for specific VMs |
| `osc_read_volumes` | List block storage volumes |
| `osc_read_nets` | List VPCs/Nets |
| `osc_read_subnets` | List Subnets |
| `osc_read_public_ips` | List Public IP addresses |
| `osc_read_security_groups` | List Security Groups and rules |
| `osc_read_images` | List machine images (OMIs) |
| `osc_read_api_logs` | Query API access logs |
| `osc_read_quotas` | List account quotas and usage |

## Example Queries

Once connected to an MCP client like Claude:

- "List all running VMs in my Outscale account"
- "Show me the security group rules for my web server"
- "What volumes are attached to VM i-12345678?"
- "Check my account quotas for VMs"
- "Show API calls that failed in the last hour"

## License

MIT
