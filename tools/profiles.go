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

// RegisterProfileTools registers profile management tools
func RegisterProfileTools(server *mcp.Server, cli *tailscale.CLI) {
	// Switch profile tool
	server.AddTool(
		&mcp.Tool{
			Name:        "switch_profile",
			Description: "Switch to a different Tailscale profile",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"profile": {Type: "string", Description: "Name of the profile to switch to"},
				},
				Required: []string{"profile"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params struct {
				Profile string `json:"profile"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			err := cli.SwitchProfile(params.Profile)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to switch to profile '%s': %v", params.Profile, err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Successfully switched to profile '%s'", params.Profile)},
				},
			}, nil
		}),
	)

	// List profiles tool
	server.AddTool(
		&mcp.Tool{
			Name:        "list_profiles",
			Description: "List all available Tailscale profiles",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			profiles, err := cli.ListProfiles()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error listing profiles: %v", err)},
					},
				}, nil
			}

			if len(profiles) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "No profiles found"},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Available Tailscale profiles:\n\n")
			for _, profile := range profiles {
				if profile.Active {
					result.WriteString(fmt.Sprintf("* %s (active)\n", profile.Name))
				} else {
					result.WriteString(fmt.Sprintf("  %s\n", profile.Name))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Get current profile tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_current_profile",
			Description: "Get the currently active Tailscale profile",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			profiles, err := cli.ListProfiles()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting current profile: %v", err)},
					},
				}, nil
			}

			for _, profile := range profiles {
				if profile.Active {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Current active profile: %s", profile.Name)},
						},
					}, nil
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No active profile found"},
				},
			}, nil
		}),
	)
}