# Tailscale MCP Server

A Model Context Protocol (MCP) server for managing Tailscale through a standardized interface. This server provides comprehensive tools for Tailscale network management, device control, and configuration.

## Features

### Authentication & Profile Management
- Switch between multiple Tailscale profiles
- List available profiles with active status
- Login/logout functionality
- Get current profile details

### Device Management
- List all devices in the network
- Get detailed device information
- Device authorization (with admin console guidance)
- Device removal and key expiry management
- Device tagging for ACL targeting

### Network Operations
- Get current network status and connectivity
- Connect/disconnect from Tailscale network
- Ping peers for connectivity testing
- Comprehensive status reporting

### Routing & Exit Nodes
- Manage exit nodes (set/clear)
- List available exit nodes
- Advertise subnet routes
- Accept routes from peers
- Configure device as exit node

### ACL & Security
- Retrieve ACL configuration
- Update ACL rules
- Create and manage authentication keys
- Key revocation and listing

### DNS & System
- Get DNS configuration and MagicDNS status
- Update DNS settings
- Get Tailscale version information
- Detailed tailnet information
- Preferences management

## Installation

### Prerequisites
- Go 1.21 or later
- Tailscale installed and configured on your system
- Access to `tailscale` CLI command

### Build from Source

```bash
# Clone the repository
git clone https://github.com/phildougherty/go-tailscale-mcp.git
cd go-tailscale-mcp

# Install dependencies
go mod download

# Build the server
go build -o tailscale-mcp

# Run the server
./tailscale-mcp
```

## Usage

The MCP server communicates via stdio, making it compatible with any MCP client. The server can be integrated with:

- Claude Desktop App
- MCP CLI tools
- Custom MCP clients

### Configuration for Claude Desktop

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "tailscale": {
      "command": "/path/to/tailscale-mcp"
    }
  }
}
```

## Available Tools

### Profile Management
- `switch_profile` - Switch between Tailscale accounts
- `list_profiles` - List all available profiles
- `get_current_profile` - Get current profile details

### Device Operations
- `list_devices` - List all network devices with details
- `get_device` - Get specific device information
- `ping_device` - Ping a device on your network

### Network Control
- `status` - Get comprehensive network status
- `connect` - Connect with advanced options
- `disconnect` - Disconnect but stay logged in
- `logout` - Complete logout from Tailscale
- `version` - Get version information

### Routing & Exit Nodes
- `set_exit_node` - Route traffic through specific node
- `clear_exit_node` - Stop using exit node
- `list_exit_nodes` - See available exit nodes
- `advertise_routes` - Share subnet routes
- `accept_routes` - Control route acceptance

### System Information
- `get_ip` - Get Tailscale IP addresses
- `get_preferences` - View all preferences
- `health_check` - Network health assessment

## Example Commands and Prompts

### Basic Status and Information

```
"What's my Tailscale status?"
"Show me all devices on my Tailscale network"
"Get the IP address for my device named 'laptop'"
"Check my Tailscale network health"
"What version of Tailscale am I running?"
```

### Profile Management

```
"List my Tailscale profiles"
"Switch to my work Tailscale profile"
"What's my current Tailscale profile?"
"Show me which Tailscale account is active"
```

### Device Operations

```
"Ping my server named 'homelab'"
"Get details about the device 'raspberrypi'"
"Show me all online devices"
"List devices with their IP addresses"
"Ping 100.64.0.1 with 10 packets"
```

### Network Connection Management

```
"Connect to Tailscale"
"Connect to Tailscale with hostname 'my-laptop' and accept routes"
"Disconnect from Tailscale network"
"Logout from my Tailscale account"
"Bring up Tailscale with SSH enabled"
```

### Exit Node Management

```
"List available exit nodes"
"Set my exit node to 'us-west-server'"
"Clear my current exit node"
"Show me which devices can be exit nodes"
"Route my traffic through the Singapore node"
```

### Routing Configuration

```
"Advertise the route 192.168.1.0/24"
"Accept routes from other devices"
"Stop accepting routes"
"Advertise my local subnet 10.0.0.0/8"
"Show advertised routes in my network"
```

### System and Preferences

```
"Show my Tailscale preferences"
"Get my current Tailscale settings"
"What are my DNS settings?"
"Show detailed network information"
```

### Advanced Examples

```
"Connect to Tailscale with auth key ABC123 and advertise 192.168.1.0/24"
"Set up this machine as an exit node and advertise it"
"Show me all devices that are currently offline"
"Ping all my devices to check connectivity"
"Get comprehensive details about my tailnet"
```

### Troubleshooting Commands

```
"Why is my Tailscale not working?"
"Check health of my Tailscale connection"
"Show me any network warnings"
"Is my device connected to Tailscale?"
"What's blocking my connection?"
```

## Architecture

The server is built using:
- **Go MCP SDK**: Official Model Context Protocol SDK for Go
- **Tailscale CLI**: Primary interface for Tailscale operations
- **Modular Design**: Tools organized by functionality

### Project Structure

```
go-tailscale-mcp/
├── main.go              # Entry point
├── server/
│   └── server.go        # MCP server setup
├── tools/
│   ├── profiles.go      # Profile management tools
│   ├── devices.go       # Device operation tools
│   ├── network.go       # Network control tools
│   ├── routing.go       # Routing and exit node tools
│   └── system.go        # System information tools
└── tailscale/
    ├── cli.go           # CLI wrapper
    └── types.go         # Type definitions
```

## Limitations

Some operations require Tailscale admin API access with an API key:
- Device authorization/removal
- ACL management
- Authentication key operations
- Advanced DNS configuration

For these operations, the server provides guidance on using the Tailscale admin console.

## Development

### Running Tests
```bash
go test ./...
```

### Adding New Tools

1. Create tool registration in appropriate file under `tools/`
2. Add CLI wrapper method if needed in `tailscale/cli.go`
3. Update types in `tailscale/types.go` if required
4. Register the tool in `server/server.go`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details

## Acknowledgments

- Built with the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Inspired by the TypeScript [Tailscale MCP](https://github.com/HexSleeves/tailscale-mcp) implementation

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/phildougherty/go-tailscale-mcp/issues)
- Check Tailscale documentation at [tailscale.com/kb](https://tailscale.com/kb)