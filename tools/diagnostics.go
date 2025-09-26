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

func RegisterDiagnosticTools(server *mcp.Server, cli *tailscale.CLI) {
	// netcheck tool
	server.AddTool(
		&mcp.Tool{
			Name:        "netcheck",
			Description: "Analyze network conditions and connectivity",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"verbose": {
						Type:        "boolean",
						Description: "Show detailed output (optional)",
					},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Verbose bool `json:"verbose"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			cmdArgs := []string{"netcheck"}
			if params.Verbose {
				cmdArgs = append(cmdArgs, "--verbose")
			}

			output, err := cli.Execute(cmdArgs...)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error running netcheck: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// whois tool
	server.AddTool(
		&mcp.Tool{
			Name:        "whois",
			Description: "Show machine and user info for a Tailscale IP",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"ip": {
						Type:        "string",
						Description: "Tailscale IP address (v4 or v6) to look up",
					},
				},
				Required: []string{"ip"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				IP string `json:"ip"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if params.IP == "" {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "IP address is required"},
					},
				}, nil
			}

			output, err := cli.Execute("whois", params.IP)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error running whois: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// bugreport tool
	server.AddTool(
		&mcp.Tool{
			Name:        "bugreport",
			Description: "Generate a shareable identifier for diagnosing issues",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"note": {
						Type:        "string",
						Description: "Optional note to include with the bug report",
					},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Note string `json:"note"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			cmdArgs := []string{"bugreport"}
			if params.Note != "" {
				cmdArgs = append(cmdArgs, "--note", params.Note)
			}

			output, err := cli.Execute(cmdArgs...)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error generating bugreport: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// serve_status tool
	server.AddTool(
		&mcp.Tool{
			Name:        "serve_status",
			Description: "Show status of Tailscale serve and funnel configurations",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"json": {
						Type:        "boolean",
						Description: "Output in JSON format (optional)",
					},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				JSON bool `json:"json"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			cmdArgs := []string{"serve", "status"}
			if params.JSON {
				cmdArgs = append(cmdArgs, "--json")
			}

			output, err := cli.Execute(cmdArgs...)
			if err != nil {
				// Check if serve is not configured
				if strings.Contains(err.Error(), "no serve config") || strings.Contains(output, "no serve config") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: "No serve configurations found"},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting serve status: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// funnel_status tool
	server.AddTool(
		&mcp.Tool{
			Name:        "funnel_status",
			Description: "Show status of Tailscale funnel configurations",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"json": {
						Type:        "boolean",
						Description: "Output in JSON format (optional)",
					},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				JSON bool `json:"json"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			cmdArgs := []string{"funnel", "status"}
			if params.JSON {
				cmdArgs = append(cmdArgs, "--json")
			}

			output, err := cli.Execute(cmdArgs...)
			if err != nil {
				// Check if funnel is not configured
				if strings.Contains(err.Error(), "no funnel config") || strings.Contains(output, "no funnel config") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: "No funnel configurations found"},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting funnel status: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// lock_status tool
	server.AddTool(
		&mcp.Tool{
			Name:        "lock_status",
			Description: "Show tailnet lock status and signing keys",
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			output, err := cli.Execute("lock", "status")
			if err != nil {
				// Check if lock is not enabled
				if strings.Contains(err.Error(), "not enabled") || strings.Contains(output, "not enabled") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: "Tailnet lock is not enabled"},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting lock status: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// lock_sign tool
	server.AddTool(
		&mcp.Tool{
			Name:        "lock_sign",
			Description: "Sign a node key and generate a signature for tailnet lock",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"node_key": {
						Type:        "string",
						Description: "Node key to sign (e.g., nodekey:abcd1234...)",
					},
				},
				Required: []string{"node_key"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				NodeKey string `json:"node_key"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if params.NodeKey == "" {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Node key is required"},
					},
				}, nil
			}

			output, err := cli.Execute("lock", "sign", params.NodeKey)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error signing node key: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// dns_status tool
	server.AddTool(
		&mcp.Tool{
			Name:        "dns_status",
			Description: "Diagnose the internal DNS forwarder",
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			output, err := cli.Execute("dns", "status")
			if err != nil {
				// Some systems may not have the DNS forwarder enabled
				if strings.Contains(err.Error(), "not running") || strings.Contains(output, "not running") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: "DNS forwarder is not running on this system"},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting DNS status: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)

	// nc tool
	server.AddTool(
		&mcp.Tool{
			Name:        "nc",
			Description: "Test connectivity to a specific port on a Tailscale host",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"host": {
						Type:        "string",
						Description: "Tailscale hostname or IP address",
					},
					"port": {
						Type:        "number",
						Description: "Port number to connect to",
					},
					"timeout": {
						Type:        "number",
						Description: "Connection timeout in seconds (optional, default 5)",
					},
				},
				Required: []string{"host", "port"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Host    string  `json:"host"`
				Port    float64 `json:"port"`
				Timeout float64 `json:"timeout"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if params.Host == "" {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Host is required"},
					},
				}, nil
			}

			port := int(params.Port)
			if port == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Port must be a valid number"},
					},
				}, nil
			}

			cmdArgs := []string{"nc"}

			// Add timeout if specified
			timeout := params.Timeout
			if timeout == 0 {
				timeout = 5
			}
			cmdArgs = append(cmdArgs, "--timeout", fmt.Sprintf("%ds", int(timeout)))

			// Add host and port
			cmdArgs = append(cmdArgs, params.Host, fmt.Sprintf("%d", port))

			output, err := cli.Execute(cmdArgs...)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Connection refused to %s:%d", params.Host, port)},
						},
					}, nil
				}
				if strings.Contains(err.Error(), "timeout") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Connection timeout to %s:%d", params.Host, port)},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to connect: %v", err)},
					},
				}, nil
			}

			// If connection succeeded
			result := fmt.Sprintf("Successfully connected to %s:%d", params.Host, port)
			if output != "" {
				result += "\n" + output
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil
		}),
	)
}