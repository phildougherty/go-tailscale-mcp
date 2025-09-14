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

// RegisterNetworkTools registers network operation tools
func RegisterNetworkTools(server *mcp.Server, cli *tailscale.CLI) {
	// Enhanced status tool
	server.AddTool(
		&mcp.Tool{
			Name:        "status",
			Description: "Get comprehensive Tailscale network status",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting status: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("=== Tailscale Network Status ===\n\n")
			result.WriteString(fmt.Sprintf("Backend State: %s\n", status.BackendState))

			if status.CurrentTailnet != nil {
				result.WriteString(fmt.Sprintf("Tailnet: %s\n", status.CurrentTailnet.Name))
				result.WriteString(fmt.Sprintf("Magic DNS: %v\n", status.CurrentTailnet.MagicDNSEnabled))
				if status.CurrentTailnet.MagicDNSEnabled {
					result.WriteString(fmt.Sprintf("DNS Suffix: %s\n", status.CurrentTailnet.MagicDNSSuffix))
				}
			}

			if status.Self != nil {
				result.WriteString(fmt.Sprintf("\n=== Your Device ===\n"))
				result.WriteString(fmt.Sprintf("Name: %s\n", status.Self.HostName))
				result.WriteString(fmt.Sprintf("OS: %s\n", status.Self.OS))
				result.WriteString(fmt.Sprintf("Online: %v\n", status.Self.Online))
				result.WriteString(fmt.Sprintf("Active: %v\n", status.Self.Active))
				if len(status.Self.TailscaleIPs) > 0 {
					result.WriteString(fmt.Sprintf("Tailscale IPs: %s\n", strings.Join(status.Self.TailscaleIPs, ", ")))
				}
				if status.Self.ExitNode {
					result.WriteString("Acting as Exit Node: Yes\n")
				}
				if status.Self.ExitNodeOption {
					result.WriteString("Available as Exit Node: Yes\n")
				}
			}

			peerCount := len(status.Peer)
			result.WriteString(fmt.Sprintf("\n=== Network Peers ===\n"))
			result.WriteString(fmt.Sprintf("Total peers: %d\n", peerCount))

			if peerCount > 0 {
				onlineCount := 0
				exitNodeCount := 0
				for _, peer := range status.Peer {
					if peer.Online {
						onlineCount++
					}
					if peer.ExitNodeOption {
						exitNodeCount++
					}
				}
				result.WriteString(fmt.Sprintf("Online peers: %d\n", onlineCount))
				result.WriteString(fmt.Sprintf("Available exit nodes: %d\n", exitNodeCount))
			}

			if len(status.Health) > 0 {
				result.WriteString(fmt.Sprintf("\n=== Health Issues ===\n"))
				for _, issue := range status.Health {
					result.WriteString(fmt.Sprintf("• %s\n", issue))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Connect with options tool
	server.AddTool(
		&mcp.Tool{
			Name:        "connect",
			Description: "Connect to Tailscale with optional configuration",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"authkey":        {Type: "string", Description: "Authentication key (optional)"},
					"accept_routes":  {Type: "boolean", Description: "Accept routes from peers (optional)"},
					"advertise_exit": {Type: "boolean", Description: "Advertise as exit node (optional)"},
					"hostname":       {Type: "string", Description: "Set custom hostname (optional)"},
					"ssh":           {Type: "boolean", Description: "Enable SSH server (optional)"},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				AuthKey        string `json:"authkey"`
				AcceptRoutes   *bool  `json:"accept_routes"`
				AdvertiseExit  *bool  `json:"advertise_exit"`
				Hostname       string `json:"hostname"`
				SSH           *bool  `json:"ssh"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			options := make(map[string]string)

			if params.AcceptRoutes != nil {
				if *params.AcceptRoutes {
					options["accept-routes"] = "true"
				} else {
					options["accept-routes"] = "false"
				}
			}

			if params.AdvertiseExit != nil && *params.AdvertiseExit {
				options["advertise-exit-node"] = "true"
			}

			if params.Hostname != "" {
				options["hostname"] = params.Hostname
			}

			if params.SSH != nil {
				if *params.SSH {
					options["ssh"] = "true"
				} else {
					options["ssh"] = "false"
				}
			}

			err := cli.Login(params.AuthKey, options)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to connect: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Successfully connected to Tailscale")
			if params.AuthKey != "" {
				result.WriteString(" with provided auth key")
			}
			if len(options) > 0 {
				result.WriteString(" with options:\n")
				for key, value := range options {
					result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
				}
			} else {
				result.WriteString("\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Disconnect tool
	server.AddTool(
		&mcp.Tool{
			Name:        "disconnect",
			Description: "Disconnect from Tailscale network (stays logged in)",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			err := cli.Down()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to disconnect: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Disconnected from Tailscale network (still logged in)"},
				},
			}, nil
		}),
	)

	// Logout tool
	server.AddTool(
		&mcp.Tool{
			Name:        "logout",
			Description: "Logout from Tailscale completely",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			err := cli.Logout()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to logout: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Logged out from Tailscale"},
				},
			}, nil
		}),
	)

	// Version tool
	server.AddTool(
		&mcp.Tool{
			Name:        "version",
			Description: "Get Tailscale version information",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			version, err := cli.Version()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting version: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Tailscale Version Information:\n%s", version)},
				},
			}, nil
		}),
	)
}