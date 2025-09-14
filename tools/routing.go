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

// RegisterRoutingTools registers routing and exit node tools
func RegisterRoutingTools(server *mcp.Server, cli *tailscale.CLI) {
	// Set exit node tool
	server.AddTool(
		&mcp.Tool{
			Name:        "set_exit_node",
			Description: "Set a specific exit node for routing internet traffic",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"node": {Type: "string", Description: "Exit node hostname, device name, or Tailscale IP"},
				},
				Required: []string{"node"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Node string `json:"node"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			err := cli.SetExitNode(params.Node)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to set exit node '%s': %v", params.Node, err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Successfully set exit node to '%s'. Internet traffic will now route through this device.", params.Node)},
				},
			}, nil
		}),
	)

	// Clear exit node tool
	server.AddTool(
		&mcp.Tool{
			Name:        "clear_exit_node",
			Description: "Clear the current exit node and route traffic directly",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			err := cli.ClearExitNode()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to clear exit node: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Cleared exit node. Internet traffic will now route directly from this device."},
				},
			}, nil
		}),
	)

	// List exit nodes tool
	server.AddTool(
		&mcp.Tool{
			Name:        "list_exit_nodes",
			Description: "List all available exit nodes in the network",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			status, err := cli.Status()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting exit node list: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Available Exit Nodes:\n\n")

			exitNodesFound := false

			// Check self device
			if status.Self != nil && status.Self.ExitNodeOption {
				result.WriteString("Your Device:\n")
				result.WriteString(fmt.Sprintf("  %s (%s) - %s\n", status.Self.HostName, strings.Join(status.Self.TailscaleIPs, ", "), status.Self.OS))
				if status.Self.ExitNode {
					result.WriteString("    Currently active as your exit node\n")
				}
				result.WriteString("\n")
				exitNodesFound = true
			}

			// Check peer devices
			for _, peer := range status.Peer {
				if peer.ExitNodeOption {
					if !exitNodesFound {
						// First peer exit node, add header if no self device was an exit node
						result.WriteString("Network Exit Nodes:\n")
					}

					onlineStatus := "Online"
					if !peer.Online {
						onlineStatus = "Offline"
					}

					result.WriteString(fmt.Sprintf("  %s (%s) - %s [%s]\n", peer.HostName, strings.Join(peer.TailscaleIPs, ", "), peer.OS, onlineStatus))
					if peer.ExitNode {
						result.WriteString("    Currently active as your exit node\n")
					}
					exitNodesFound = true
				}
			}

			if !exitNodesFound {
				result.WriteString("No exit nodes available in the network.\n")
				result.WriteString("Exit nodes must be explicitly enabled on devices to appear here.\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Advertise routes tool
	server.AddTool(
		&mcp.Tool{
			Name:        "advertise_routes",
			Description: "Advertise subnet routes to other devices in the network",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"routes": {
						Type: "array",
						Items: &jsonschema.Schema{Type: "string"},
						Description: "List of subnet routes to advertise (e.g., ['192.168.1.0/24', '10.0.0.0/8'])",
					},
				},
				Required: []string{"routes"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Routes []string `json:"routes"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if len(params.Routes) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "No routes specified. Please provide at least one route to advertise."},
					},
				}, nil
			}

			err := cli.AdvertiseRoutes(params.Routes)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to advertise routes: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Successfully advertising routes: %s\n\nNote: Routes may need approval in the Tailscale admin console.", strings.Join(params.Routes, ", "))},
				},
			}, nil
		}),
	)

	// Accept routes tool
	server.AddTool(
		&mcp.Tool{
			Name:        "accept_routes",
			Description: "Enable or disable accepting subnet routes from peers",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"accept": {Type: "boolean", Description: "Whether to accept routes from peers"},
				},
				Required: []string{"accept"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Accept bool `json:"accept"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			err := cli.AcceptRoutes(params.Accept)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to update route acceptance: %v", err)},
					},
				}, nil
			}

			status := "disabled"
			message := "This device will no longer accept subnet routes from peers."
			if params.Accept {
				status = "enabled"
				message = "This device will now accept subnet routes advertised by peers."
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Route acceptance %s. %s", status, message)},
				},
			}, nil
		}),
	)
}