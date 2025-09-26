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
					"profile": {Type: "string", Description: "Profile ID, email address, or tailnet name to switch to"},
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

			// Get list of profiles to find the right one
			profiles, err := cli.ListProfiles()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to list profiles: %v", err)},
					},
				}, nil
			}

			// Find matching profile by ID, account, or tailnet
			var targetProfile *tailscale.Profile
			inputLower := strings.ToLower(params.Profile)

			for _, p := range profiles {
				profile := p // Create a copy for pointer
				// Exact match on ID
				if strings.EqualFold(p.ID, params.Profile) {
					targetProfile = &profile
					break
				}
				// Match on account email
				if strings.EqualFold(p.Account, params.Profile) {
					targetProfile = &profile
					break
				}
				// Match on tailnet
				if strings.EqualFold(p.Tailnet, params.Profile) {
					targetProfile = &profile
					break
				}
				// Partial match on account or tailnet
				if strings.Contains(strings.ToLower(p.Account), inputLower) ||
					strings.Contains(strings.ToLower(p.Tailnet), inputLower) {
					if targetProfile == nil {
						targetProfile = &profile
					} else {
						// Multiple matches, need to be more specific
						return &mcp.CallToolResult{
							Content: []mcp.Content{
								&mcp.TextContent{Text: fmt.Sprintf("Multiple profiles match '%s'. Please be more specific or use the profile ID.", params.Profile)},
							},
						}, nil
					}
				}
			}

			if targetProfile == nil {
				// List available profiles to help the user
				var profileList strings.Builder
				profileList.WriteString(fmt.Sprintf("Profile '%s' not found. Available profiles:\n", params.Profile))
				for _, p := range profiles {
					profileList.WriteString(fmt.Sprintf("  ID: %s, Account: %s\n", p.ID, p.Account))
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: profileList.String()},
					},
				}, nil
			}

			// Check if already on this profile
			if targetProfile.Active {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Already on profile %s (%s)", targetProfile.ID, targetProfile.Account)},
					},
				}, nil
			}

			// Switch using the profile ID
			err = cli.SwitchProfile(targetProfile.ID)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to switch to profile '%s': %v", targetProfile.Account, err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Successfully switched to profile:\n  ID: %s\n  Account: %s\n  Tailnet: %s",
						targetProfile.ID, targetProfile.Account, targetProfile.Tailnet)},
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
			result.WriteString(fmt.Sprintf("%-6s %-30s %-30s\n", "ID", "Tailnet", "Account"))
			for _, profile := range profiles {
				marker := " "
				if profile.Active {
					marker = "*"
				}
				result.WriteString(fmt.Sprintf("%-6s %-30s %-30s%s\n",
					profile.ID, profile.Tailnet, profile.Account, marker))
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
							&mcp.TextContent{Text: fmt.Sprintf("Current active profile:\n  ID: %s\n  Tailnet: %s\n  Account: %s",
								profile.ID, profile.Tailnet, profile.Account)},
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

	// Add new profile tool
	server.AddTool(
		&mcp.Tool{
			Name:        "add_profile",
			Description: "Add a new Tailscale profile by logging in to a different account",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Start the login process for a new profile
			output, err := cli.LoginNewProfile()
			if err != nil {
				// Check if it's because we need to specify a different account
				if strings.Contains(err.Error(), "already logged in") || strings.Contains(output, "already logged in") {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: "You're already logged in to a profile. To add a new profile:\n" +
								"1. First logout from the current profile with 'tailscale logout'\n" +
								"2. Then login with your new account\n" +
								"3. The new profile will be automatically added\n\n" +
								"Alternatively, use the auth URL that should appear when running 'tailscale login' again."},
						},
					}, nil
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Failed to start login process: %v", err)},
					},
				}, nil
			}

			// Extract auth URL if present
			if strings.Contains(output, "https://") {
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "https://login.tailscale.com/") {
						return &mcp.CallToolResult{
							Content: []mcp.Content{
								&mcp.TextContent{Text: fmt.Sprintf("To add a new profile, authenticate at:\n%s\n\n" +
									"After authentication, the new profile will be automatically added.", strings.TrimSpace(line))},
							},
						}, nil
					}
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, nil
		}),
	)
}