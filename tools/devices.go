package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/tailscale"
)

// RegisterDeviceTools registers device operation tools
func RegisterDeviceTools(server *mcp.Server, cli *tailscale.CLI) {
	// List devices tool
	server.AddTool(
		&mcp.Tool{
			Name:        "list_devices",
			Description: "List all devices in the Tailscale network",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting device list: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Tailscale Network Devices:\n\n")

			// Show self device first
			if status.Self != nil {
				result.WriteString("Your Device:\n")
				result.WriteString(fmt.Sprintf("  Name: %s\n", status.Self.HostName))
				result.WriteString(fmt.Sprintf("  OS: %s\n", status.Self.OS))
				result.WriteString(fmt.Sprintf("  Online: %v\n", status.Self.Online))
				if len(status.Self.TailscaleIPs) > 0 {
					result.WriteString(fmt.Sprintf("  IPs: %s\n", strings.Join(status.Self.TailscaleIPs, ", ")))
				}
				if status.Self.ExitNode {
					result.WriteString("  Role: Exit Node\n")
				}
				result.WriteString("\n")
			}

			// Show peer devices
			if len(status.Peer) > 0 {
				result.WriteString("Other Devices:\n")
				for _, peer := range status.Peer {
					result.WriteString(fmt.Sprintf("  Name: %s\n", peer.HostName))
					result.WriteString(fmt.Sprintf("  OS: %s\n", peer.OS))
					result.WriteString(fmt.Sprintf("  Online: %v\n", peer.Online))
					if len(peer.TailscaleIPs) > 0 {
						result.WriteString(fmt.Sprintf("  IPs: %s\n", strings.Join(peer.TailscaleIPs, ", ")))
					}
					if peer.ExitNode {
						result.WriteString("  Role: Exit Node\n")
					}
					if peer.ExitNodeOption {
						result.WriteString("  Available as Exit Node\n")
					}
					result.WriteString("\n")
				}
			} else {
				result.WriteString("No other devices found in network\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Get specific device tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_device",
			Description: "Get detailed information about a specific device",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device": {Type: "string", Description: "Device name or hostname to get information for"},
				},
				Required: []string{"device"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Device string `json:"device"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting device information: %v", err)},
					},
				}, nil
			}

			var targetPeer *tailscale.PeerStatus
			deviceName := strings.ToLower(params.Device)

			// Check if it's the self device
			if status.Self != nil && strings.ToLower(status.Self.HostName) == deviceName {
				var result strings.Builder
				result.WriteString(fmt.Sprintf("Device Details: %s (Your Device)\n\n", status.Self.HostName))
				result.WriteString(fmt.Sprintf("Hostname: %s\n", status.Self.HostName))
				result.WriteString(fmt.Sprintf("DNS Name: %s\n", status.Self.DNSName))
				result.WriteString(fmt.Sprintf("OS: %s\n", status.Self.OS))
				result.WriteString(fmt.Sprintf("Online: %v\n", status.Self.Online))
				result.WriteString(fmt.Sprintf("Active: %v\n", status.Self.Active))
				if len(status.Self.TailscaleIPs) > 0 {
					result.WriteString(fmt.Sprintf("Tailscale IPs: %s\n", strings.Join(status.Self.TailscaleIPs, ", ")))
				}
				if len(status.Self.AllowedIPs) > 0 {
					result.WriteString(fmt.Sprintf("Allowed IPs: %s\n", strings.Join(status.Self.AllowedIPs, ", ")))
				}
				if status.Self.ExitNode {
					result.WriteString("Role: Exit Node\n")
				}
				if status.Self.ExitNodeOption {
					result.WriteString("Available as Exit Node: Yes\n")
				}
				result.WriteString(fmt.Sprintf("Public Key: %s\n", status.Self.PublicKey))
				result.WriteString(fmt.Sprintf("Last Seen: %s\n", status.Self.LastSeen.Format("2006-01-02 15:04:05")))

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: result.String()},
					},
				}, nil
			}

			// Look for the device in peers
			for _, peer := range status.Peer {
				if strings.ToLower(peer.HostName) == deviceName || strings.ToLower(peer.DNSName) == deviceName {
					targetPeer = peer
					break
				}
			}

			if targetPeer == nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Device '%s' not found in network", params.Device)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString(fmt.Sprintf("Device Details: %s\n\n", targetPeer.HostName))
			result.WriteString(fmt.Sprintf("Hostname: %s\n", targetPeer.HostName))
			result.WriteString(fmt.Sprintf("DNS Name: %s\n", targetPeer.DNSName))
			result.WriteString(fmt.Sprintf("OS: %s\n", targetPeer.OS))
			result.WriteString(fmt.Sprintf("Online: %v\n", targetPeer.Online))
			result.WriteString(fmt.Sprintf("Active: %v\n", targetPeer.Active))
			if len(targetPeer.TailscaleIPs) > 0 {
				result.WriteString(fmt.Sprintf("Tailscale IPs: %s\n", strings.Join(targetPeer.TailscaleIPs, ", ")))
			}
			if len(targetPeer.AllowedIPs) > 0 {
				result.WriteString(fmt.Sprintf("Allowed IPs: %s\n", strings.Join(targetPeer.AllowedIPs, ", ")))
			}
			if targetPeer.ExitNode {
				result.WriteString("Role: Exit Node\n")
			}
			if targetPeer.ExitNodeOption {
				result.WriteString("Available as Exit Node: Yes\n")
			}
			if len(targetPeer.Tags) > 0 {
				result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(targetPeer.Tags, ", ")))
			}
			result.WriteString(fmt.Sprintf("Public Key: %s\n", targetPeer.PublicKey))
			result.WriteString(fmt.Sprintf("Last Seen: %s\n", targetPeer.LastSeen.Format("2006-01-02 15:04:05")))
			result.WriteString(fmt.Sprintf("RX Bytes: %d\n", targetPeer.RxBytes))
			result.WriteString(fmt.Sprintf("TX Bytes: %d\n", targetPeer.TxBytes))

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Ping device tool
	server.AddTool(
		&mcp.Tool{
			Name:        "ping_device",
			Description: "Ping a specific device in the Tailscale network",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device": {Type: "string", Description: "Device name, hostname, or IP address to ping"},
					"count":  {Type: "integer", Description: "Number of pings to send (default: 4)"},
				},
				Required: []string{"device"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Device string `json:"device"`
				Count  int    `json:"count"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Default count to 4 if not specified
			if params.Count <= 0 {
				params.Count = 4
			}

			result, err := cli.Ping(params.Device, params.Count)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to ping %s: %v", params.Device, err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Ping results for %s:\n\n%s", params.Device, result)},
				},
			}, nil
		}),
	)
}

// RegisterDeviceToolsWithAPI registers device operation tools with API client support
func RegisterDeviceToolsWithAPI(server *mcp.Server, cli *tailscale.CLI, api *tailscale.APIClient) {
	// Register all existing CLI-based tools first
	RegisterDeviceTools(server, cli)

	// Authorize device tool (API-enhanced)
	server.AddTool(
		&mcp.Tool{
			Name:        "authorize_device",
			Description: "Authorize a device in the Tailscale network",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device_id": {
						Type:        "string",
						Description: "Device ID to authorize",
					},
				},
				Required: []string{"device_id"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				DeviceID string `json:"device_id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Try API first if available
			if api != nil && api.IsAvailable() {
				if err := api.AuthorizeDevice(params.DeviceID); err != nil {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Error authorizing device via API: %v", err)},
						},
					}, nil
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Device %s authorized successfully via API.", params.DeviceID)},
					},
				}, nil
			}

			// Fallback to CLI (if implemented)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "API client not configured. Device authorization requires API access. Please set TAILSCALE_API_KEY environment variable."},
				},
			}, nil
		}),
	)

	// Delete device tool (API-enhanced)
	server.AddTool(
		&mcp.Tool{
			Name:        "delete_device",
			Description: "Remove a device from the Tailscale network",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device_id": {
						Type:        "string",
						Description: "Device ID to remove",
					},
				},
				Required: []string{"device_id"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				DeviceID string `json:"device_id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Try API first if available
			if api != nil && api.IsAvailable() {
				if err := api.DeleteDevice(params.DeviceID); err != nil {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Error deleting device via API: %v", err)},
						},
					}, nil
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Device %s deleted successfully via API.", params.DeviceID)},
					},
				}, nil
			}

			// Fallback to CLI (if implemented)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "API client not configured. Device deletion requires API access. Please set TAILSCALE_API_KEY environment variable."},
				},
			}, nil
		}),
	)

	// Set device tags tool (API-enhanced)
	server.AddTool(
		&mcp.Tool{
			Name:        "set_device_tags",
			Description: "Set tags for a device",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device_id": {
						Type:        "string",
						Description: "Device ID to set tags for",
					},
					"tags": {
						Type: "array",
						Items: &jsonschema.Schema{Type: "string"},
						Description: "Tags to set for the device",
					},
				},
				Required: []string{"device_id", "tags"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				DeviceID string   `json:"device_id"`
				Tags     []string `json:"tags"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Try API first if available
			if api != nil && api.IsAvailable() {
				if err := api.SetDeviceTags(params.DeviceID, params.Tags); err != nil {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Error setting device tags via API: %v", err)},
						},
					}, nil
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Tags set successfully for device %s: %s", params.DeviceID, strings.Join(params.Tags, ", "))},
					},
				}, nil
			}

			// Fallback to CLI (if implemented)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "API client not configured. Setting device tags requires API access. Please set TAILSCALE_API_KEY environment variable."},
				},
			}, nil
		}),
	)
}