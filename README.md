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

### Option 1: Environment Variables (Single Account)

Set the following environment variables with your Outscale credentials:

```bash
export OSC_ACCESS_KEY="your-access-key"
export OSC_SECRET_KEY="your-secret-key"
export OSC_REGION="eu-west-2"  # Optional, defaults to eu-west-2
```

### Option 2: Configuration File (Multiple Accounts)

Create a configuration file at `~/.osc/config.json` (or specify a custom path with `OSC_CONFIG_FILE`):

```json
{
  "default": {
    "access_key": "your-default-access-key",
    "secret_key": "your-default-secret-key",
    "region": "eu-west-2"
  },
  "production": {
    "access_key": "your-prod-access-key",
    "secret_key": "your-prod-secret-key",
    "region": "eu-west-2"
  },
  "development": {
    "access_key": "your-dev-access-key",
    "secret_key": "your-dev-secret-key",
    "region": "us-east-2"
  }
}
```

### Priority

1. If `OSC_ACCESS_KEY` and `OSC_SECRET_KEY` are set → used as the "default" profile
2. Otherwise → load from `OSC_CONFIG_FILE` or `~/.osc/config.json`

## Usage

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

**Single account (environment variables):**
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

**Multiple accounts (config file):**
```json
{
  "mcpServers": {
    "outscale": {
      "command": "/path/to/outscale-mcp",
      "env": {
        "OSC_CONFIG_FILE": "/path/to/config.json"
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
| `osc_list_profiles` | List all available Outscale profiles |
| `osc_check_auth` | Verify API credentials are valid |
| `osc_read_vms` | List and inspect virtual machines |
| `osc_read_vm_state` | Get detailed state for specific VMs |
| `osc_read_volumes` | List block storage volumes |
| `osc_read_nets` | List VPCs/Nets |
| `osc_read_subnets` | List Subnets |
| `osc_read_route_tables` | List Route Tables and their routes |
| `osc_read_internet_services` | List Internet Services (Internet Gateways) |
| `osc_read_nat_services` | List NAT Services and their public IPs |
| `osc_read_net_peerings` | List Net Peerings between VPCs |
| `osc_read_net_access_points` | List Net Access Points (VPC Endpoints) |
| `osc_read_public_ips` | List Public IP addresses |
| `osc_read_security_groups` | List Security Groups and rules |
| `osc_read_images` | List machine images (OMIs) |
| `osc_read_api_logs` | Query API access logs |
| `osc_read_quotas` | List account quotas and usage |
| `osc_read_load_balancers` | List Load Balancers with listeners, backends, and health checks |
| `osc_read_console_output` | Get VM boot logs (console output) |

### Profile Parameter

All tools accept an optional `profile` parameter to specify which Outscale account to use. If not specified, the default profile is used.

Example queries:
- "List all running VMs" (uses default profile)
- "List all running VMs using the production profile"
- "Show me the security group rules for profile development"

## Example Queries

Once connected to an MCP client like Claude:

- "List all running VMs in my Outscale account"
- "Show me the security group rules for my web server"
- "What volumes are attached to VM i-12345678?"
- "Check my account quotas for VMs"
- "Show API calls that failed in the last hour"
- "List all available profiles"
- "Check authentication for the production profile"

## License

MIT
