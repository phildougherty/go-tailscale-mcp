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

// RegisterSystemTools registers system information tools
func RegisterSystemTools(server *mcp.Server, cli *tailscale.CLI) {
	// Get IP tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_ip",
			Description: "Get Tailscale IP addresses for this device or a specific device",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"device": {Type: "string", Description: "Device name to get IP for (optional, defaults to this device)"},
				},
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

			ip, err := cli.IP(params.Device)
			if err != nil {
				if params.Device != "" {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Failed to get IP for device '%s': %v", params.Device, err)},
						},
					}, nil
				} else {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Failed to get IP: %v", err)},
						},
					}, nil
				}
			}

			var result string
			if params.Device != "" {
				result = fmt.Sprintf("Tailscale IP for device '%s':\n%s", params.Device, ip)
			} else {
				result = fmt.Sprintf("Your Tailscale IP addresses:\n%s", ip)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil
		}),
	)

	// Get preferences/settings tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_preferences",
			Description: "Get current Tailscale preferences and settings",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting preferences: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("=== Tailscale Preferences & Settings ===\n\n")

			// Network state
			result.WriteString(fmt.Sprintf("Connection State: %s\n", status.BackendState))

			if status.CurrentTailnet != nil {
				result.WriteString(fmt.Sprintf("Current Tailnet: %s\n", status.CurrentTailnet.Name))
				result.WriteString(fmt.Sprintf("Magic DNS Enabled: %v\n", status.CurrentTailnet.MagicDNSEnabled))
				if status.CurrentTailnet.MagicDNSEnabled {
					result.WriteString(fmt.Sprintf("Magic DNS Suffix: %s\n", status.CurrentTailnet.MagicDNSSuffix))
				}
			}

			// Device settings
			if status.Self != nil {
				result.WriteString("\n=== Device Settings ===\n")
				result.WriteString(fmt.Sprintf("Hostname: %s\n", status.Self.HostName))
				result.WriteString(fmt.Sprintf("DNS Name: %s\n", status.Self.DNSName))
				result.WriteString(fmt.Sprintf("Online: %v\n", status.Self.Online))
				result.WriteString(fmt.Sprintf("Active: %v\n", status.Self.Active))

				// Exit node settings
				if status.Self.ExitNode {
					result.WriteString("Acting as Exit Node: Yes\n")
				} else {
					result.WriteString("Acting as Exit Node: No\n")
				}

				if status.Self.ExitNodeOption {
					result.WriteString("Available as Exit Node: Yes\n")
				} else {
					result.WriteString("Available as Exit Node: No\n")
				}

				// Network information
				if len(status.Self.TailscaleIPs) > 0 {
					result.WriteString(fmt.Sprintf("Tailscale IPs: %s\n", strings.Join(status.Self.TailscaleIPs, ", ")))
				}
				if len(status.Self.AllowedIPs) > 0 {
					result.WriteString(fmt.Sprintf("Allowed IPs: %s\n", strings.Join(status.Self.AllowedIPs, ", ")))
				}

				// Tags
				if len(status.Self.Tags) > 0 {
					result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(status.Self.Tags, ", ")))
				}

				// Key information
				result.WriteString(fmt.Sprintf("Key Expiry: %s\n", status.Self.KeyExpiry.Format("2006-01-02 15:04:05")))
				if status.Self.Expired {
					result.WriteString("Key Status: EXPIRED\n")
				} else {
					result.WriteString("Key Status: Valid\n")
				}
			}

			// Exit node usage (check if using another device as exit node)
			result.WriteString("\n=== Exit Node Usage ===\n")
			usingExitNode := false
			for _, peer := range status.Peer {
				if peer.ExitNode {
					result.WriteString(fmt.Sprintf("Using Exit Node: %s (%s)\n", peer.HostName, strings.Join(peer.TailscaleIPs, ", ")))
					usingExitNode = true
					break
				}
			}
			if !usingExitNode {
				result.WriteString("Using Exit Node: None (direct routing)\n")
			}

			// Route information
			routeInfo := false
			for _, peer := range status.Peer {
				if len(peer.PrimaryRoutes) > 0 {
					if !routeInfo {
						result.WriteString("\n=== Advertised Routes ===\n")
						routeInfo = true
					}
					result.WriteString(fmt.Sprintf("%s: %s\n", peer.HostName, strings.Join(peer.PrimaryRoutes, ", ")))
				}
			}
			if !routeInfo {
				result.WriteString("\n=== Advertised Routes ===\n")
				result.WriteString("No subnet routes advertised in the network\n")
			}

			// Health status
			if len(status.Health) > 0 {
				result.WriteString("\n=== Health Issues ===\n")
				for _, issue := range status.Health {
					result.WriteString(fmt.Sprintf("• %s\n", issue))
				}
			} else {
				result.WriteString("\n=== Health Status ===\n")
				result.WriteString("No health issues detected\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Network health check tool
	server.AddTool(
		&mcp.Tool{
			Name:        "health_check",
			Description: "Check Tailscale network health and connectivity",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error performing health check: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("=== Tailscale Health Check ===\n\n")

			// Connection status
			switch status.BackendState {
			case "Running":
				result.WriteString("✓ Connection Status: Connected and running\n")
			case "NeedsLogin":
				result.WriteString("⚠ Connection Status: Needs login\n")
			case "NeedsMachineAuth":
				result.WriteString("⚠ Connection Status: Needs machine authorization\n")
			case "Stopped":
				result.WriteString("✗ Connection Status: Stopped\n")
			default:
				result.WriteString(fmt.Sprintf("? Connection Status: %s\n", status.BackendState))
			}

			// Self device status
			if status.Self != nil {
				if status.Self.Online {
					result.WriteString("✓ Device Status: Online\n")
				} else {
					result.WriteString("⚠ Device Status: Offline\n")
				}

				if status.Self.Expired {
					result.WriteString("✗ Key Status: Expired\n")
				} else {
					result.WriteString("✓ Key Status: Valid\n")
				}
			}

			// Peer connectivity
			totalPeers := len(status.Peer)
			onlinePeers := 0
			for _, peer := range status.Peer {
				if peer.Online {
					onlinePeers++
				}
			}

			result.WriteString(fmt.Sprintf("✓ Network Peers: %d total, %d online\n", totalPeers, onlinePeers))

			// DNS status
			if status.CurrentTailnet != nil {
				if status.CurrentTailnet.MagicDNSEnabled {
					result.WriteString("✓ Magic DNS: Enabled\n")
				} else {
					result.WriteString("⚠ Magic DNS: Disabled\n")
				}
			}

			// Health issues
			if len(status.Health) > 0 {
				result.WriteString(fmt.Sprintf("\n⚠ Health Issues Detected (%d):\n", len(status.Health)))
				for i, issue := range status.Health {
					result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, issue))
				}
			} else {
				result.WriteString("\n✓ No health issues detected\n")
			}

			// Overall assessment
			result.WriteString("\n=== Overall Assessment ===\n")
			if status.BackendState == "Running" && status.Self != nil && status.Self.Online && !status.Self.Expired && len(status.Health) == 0 {
				result.WriteString("✓ HEALTHY: Tailscale is functioning normally\n")
			} else if status.BackendState == "Running" && status.Self != nil && status.Self.Online {
				result.WriteString("⚠ MINOR ISSUES: Tailscale is connected but has some issues\n")
			} else {
				result.WriteString("✗ ISSUES DETECTED: Tailscale needs attention\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)
}