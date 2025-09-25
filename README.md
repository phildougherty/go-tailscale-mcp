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

### Kubernetes Operator Management (Optional)
- Manage Tailscale Kubernetes operator resources
- Create ProxyGroups, ProxyClasses, Connectors, and DNSConfigs
- Configure Tailscale Ingress and Egress services
- Requires manual operator installation first

## Installation

### Prerequisites
- Go 1.21 or later
- Tailscale installed and configured on your system
- Access to `tailscale` CLI command
- (Optional) Kubernetes cluster and kubectl configured for operator features

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
      "command": "/path/to/tailscale-mcp",
      "env": {
        "TAILSCALE_API_KEY": "tskey-api-...",
        "TAILSCALE_TAILNET": "your-email@example.com"
      }
    }
  }
}
```

To enable Kubernetes operator features, set the `ENABLE_K8S_OPERATOR` environment variable:

```json
{
  "mcpServers": {
    "tailscale": {
      "command": "/path/to/tailscale-mcp",
      "env": {
        "TAILSCALE_API_KEY": "tskey-api-...",
        "TAILSCALE_TAILNET": "your-email@example.com",
        "ENABLE_K8S_OPERATOR": "true",
        "KUBECONFIG": "/path/to/kubeconfig"
      }
    }
  }
}
```

Or using the Claude CLI:

```bash
# Add with API support
claude mcp add -s user tailscale /path/to/tailscale-mcp \
  -e TAILSCALE_API_KEY=tskey-api-... \
  -e TAILSCALE_TAILNET=your-email@example.com

# Add with Kubernetes operator support
claude mcp add -s user tailscale /path/to/tailscale-mcp \
  -e TAILSCALE_API_KEY=tskey-api-... \
  -e TAILSCALE_TAILNET=your-email@example.com \
  -e ENABLE_K8S_OPERATOR=true
```

### API Configuration

To enable full functionality including device authorization, ACL management, and auth key operations, configure the Tailscale API:

1. **Get a Tailscale API Key:**
   - Go to https://login.tailscale.com/admin/settings/keys
   - Create a new API key
   - Copy the key (starts with `tskey-api-`)

2. **Set Environment Variables:**
   ```bash
   export TAILSCALE_API_KEY="tskey-api-..."
   export TAILSCALE_TAILNET="your-email@example.com"  # or your organization domain
   ```

3. **API-Enabled Features:**
   With the API configured, you gain access to:
   - Device authorization and removal
   - ACL policy management
   - Authentication key creation and management
   - DNS configuration
   - Route approval
   - Device tagging

Without the API, the server still provides full network management through the CLI tools.

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
│   ├── system.go        # System information tools
│   ├── acl.go           # ACL management tools
│   ├── authkeys.go      # Authentication key tools
│   └── dns_api.go       # DNS API configuration tools
├── tailscale/
│   ├── cli.go           # CLI wrapper
│   ├── api.go           # Tailscale API client
│   └── types.go         # Type definitions
└── k8s/
    ├── client.go        # Kubernetes client setup
    ├── operator.go      # Operator management functions
    ├── resources.go     # Custom resource definitions
    ├── errors.go        # Error handling and types
    └── tools.go         # Kubernetes MCP tools
```

### API-Only Tools (Requires TAILSCALE_API_KEY)

#### ACL Management
- `get_acl` - Get current ACL policy
- `update_acl` - Update ACL policy with validation
- `validate_acl` - Validate ACL without applying

#### Authentication Keys
- `create_auth_key` - Create new auth key with options
- `list_auth_keys` - List all auth keys with details
- `delete_auth_key` - Delete an auth key

#### DNS API Configuration
- `get_dns_config` - Get complete DNS configuration
- `set_dns_nameservers` - Configure DNS nameservers
- `set_dns_preferences` - Enable/disable MagicDNS
- `set_dns_search_paths` - Set DNS search paths

#### Enhanced Device Operations (with API)
- `authorize_device` - Authorize pending devices (API-enabled)
- `delete_device` - Remove devices from network (API-enabled)
- `set_device_tags` - Manage device tags (API-enabled)

#### Route Management (with API)
- `approve_routes` - Approve advertised routes (API-enabled)

### Kubernetes Operator Tools (Requires ENABLE_K8S_OPERATOR=true)

**Prerequisites:**
1. Install the Tailscale Kubernetes operator first:
   ```bash
   # Using kubectl
   kubectl apply -f https://tailscale.com/install/kubernetes/operator.yaml

   # Or using Helm
   helm repo add tailscale https://pkgs.tailscale.com/helmcharts
   helm install tailscale-operator tailscale/tailscale-operator
   ```
2. Set `ENABLE_K8S_OPERATOR=true` in your MCP configuration
3. Ensure `kubectl` is configured with cluster access

#### Example Prompts

**Exposing Services to Tailnet (Ingress):**
```
"Expose my nginx service on port 80 as 'webapp' to the tailnet"
"Create a Tailscale ingress for service 'api-server' on port 8080"
"Make my grafana dashboard accessible via Tailscale at hostname 'monitoring'"
```
Use case: Access internal Kubernetes services securely via your Tailscale network without public exposure.

**Accessing External Tailnet Services (Egress):**
```
"Create an egress to access my database at 'db.tailnet' on port 5432"
"Connect to external service 'backup-server.tailnet' from inside the cluster"
"Set up egress for 'metrics-collector.tailnet' on port 9090"
```
Use case: Allow pods to connect to services running elsewhere in your tailnet.

**High Availability with ProxyGroups:**
```
"Create a ProxyGroup with 3 replicas for high availability egress"
"Deploy a ProxyGroup named 'ha-proxy' with type 'ingress' and 2 replicas"
"Scale the ProxyGroup 'production-proxy' to 5 replicas"
```
Use case: Ensure resilient connectivity with multiple proxy replicas for production workloads.

**Subnet Routing with Connectors:**
```
"Create a Connector to advertise subnet 10.0.0.0/24 to the tailnet"
"Set up a Connector as an exit node for the cluster"
"Deploy a Connector with hostname 'k8s-subnet' advertising routes 192.168.1.0/24"
```
Use case: Share cluster pod/service networks with your tailnet or route cluster traffic through Tailscale.

**DNS Configuration:**
```
"Enable MagicDNS for the cluster"
"Create a DNSConfig with MagicDNS enabled"
"Configure cluster DNS to use Tailscale MagicDNS"
```
Use case: Enable automatic DNS resolution for tailnet hostnames within the cluster.

**ProxyClass for Custom Configuration:**
```
"Create a ProxyClass named 'production' with specific labels"
"Deploy a ProxyClass for custom proxy pod configuration"
```
Use case: Define reusable proxy configurations for different environments or requirements.

## Configuration Options

### Environment Variables

- `TAILSCALE_API_KEY` - Your Tailscale API key for admin operations
- `TAILSCALE_TAILNET` - Your tailnet domain (e.g., your-email@example.com or org.domain)
- `ENABLE_K8S_OPERATOR` - Set to `true` to enable Kubernetes operator management features
- `KUBECONFIG` - Path to kubeconfig file (optional, defaults to ~/.kube/config)

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