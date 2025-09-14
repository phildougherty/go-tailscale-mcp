package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/tailscale"
)

// RegisterAuthKeyTools registers authentication key management tools
func RegisterAuthKeyTools(server *mcp.Server, api *tailscale.APIClient) {
	// Create auth key tool
	server.AddTool(
		&mcp.Tool{
			Name:        "create_auth_key",
			Description: "Create a new authentication key with specified options",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"reusable": {
						Type:        "boolean",
						Description: "Whether the key can be used multiple times (default: false)",
					},
					"ephemeral": {
						Type:        "boolean",
						Description: "Whether devices using this key are ephemeral (default: false)",
					},
					"preauthorized": {
						Type:        "boolean",
						Description: "Whether devices using this key are automatically authorized (default: false)",
					},
					"tags": {
						Type: "array",
						Items: &jsonschema.Schema{Type: "string"},
						Description: "Tags to assign to devices using this key",
					},
					"expiry_seconds": {
						Type:        "integer",
						Description: "Key expiration time in seconds (default: 3600)",
					},
				},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if api == nil || !api.IsAvailable() {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "API client not configured. Please set TAILSCALE_API_KEY environment variable."},
					},
				}, nil
			}

			var params struct {
				Reusable      *bool    `json:"reusable"`
				Ephemeral     *bool    `json:"ephemeral"`
				Preauthorized *bool    `json:"preauthorized"`
				Tags          []string `json:"tags"`
				ExpirySeconds *int     `json:"expiry_seconds"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Set defaults
			options := tailscale.AuthKeyOptions{
				Reusable:      false,
				Ephemeral:     false,
				Preauthorized: false,
				ExpirySeconds: 3600, // 1 hour default
			}

			if params.Reusable != nil {
				options.Reusable = *params.Reusable
			}
			if params.Ephemeral != nil {
				options.Ephemeral = *params.Ephemeral
			}
			if params.Preauthorized != nil {
				options.Preauthorized = *params.Preauthorized
			}
			if params.Tags != nil {
				options.Tags = params.Tags
			}
			if params.ExpirySeconds != nil {
				options.ExpirySeconds = *params.ExpirySeconds
			}

			authKey, err := api.CreateAuthKey(options)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error creating auth key: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Authentication Key Created:\n\n")
			result.WriteString(fmt.Sprintf("ID: %s\n", authKey.ID))
			result.WriteString(fmt.Sprintf("Key: %s\n", authKey.Key))
			result.WriteString(fmt.Sprintf("Created: %s\n", authKey.Created.Format("2006-01-02 15:04:05")))
			result.WriteString(fmt.Sprintf("Expires: %s\n", authKey.Expires.Format("2006-01-02 15:04:05")))
			result.WriteString(fmt.Sprintf("Reusable: %t\n", authKey.Reusable))
			result.WriteString(fmt.Sprintf("Ephemeral: %t\n", authKey.Ephemeral))
			result.WriteString(fmt.Sprintf("Preauthorized: %t\n", authKey.Preauthorized))
			if len(authKey.Tags) > 0 {
				result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(authKey.Tags, ", ")))
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// List auth keys tool
	server.AddTool(
		&mcp.Tool{
			Name:        "list_auth_keys",
			Description: "List all authentication keys",
			InputSchema: &jsonschema.Schema{Type: "object"},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if api == nil || !api.IsAvailable() {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "API client not configured. Please set TAILSCALE_API_KEY environment variable."},
					},
				}, nil
			}

			authKeys, err := api.ListAuthKeys()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error listing auth keys: %v", err)},
					},
				}, nil
			}

			if len(authKeys) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "No authentication keys found."},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("Authentication Keys:\n\n")

			for _, key := range authKeys {
				result.WriteString(fmt.Sprintf("ID: %s\n", key.ID))

				// Only show part of the key for security
				keyDisplay := key.Key
				if len(keyDisplay) > 20 {
					keyDisplay = keyDisplay[:20] + "..."
				}
				result.WriteString(fmt.Sprintf("Key: %s\n", keyDisplay))

				result.WriteString(fmt.Sprintf("Created: %s\n", key.Created.Format("2006-01-02 15:04:05")))
				result.WriteString(fmt.Sprintf("Expires: %s\n", key.Expires.Format("2006-01-02 15:04:05")))

				// Check if expired
				if time.Now().After(key.Expires) {
					result.WriteString("Status: EXPIRED\n")
				} else {
					result.WriteString("Status: Active\n")
				}

				result.WriteString(fmt.Sprintf("Reusable: %t\n", key.Reusable))
				result.WriteString(fmt.Sprintf("Ephemeral: %t\n", key.Ephemeral))
				result.WriteString(fmt.Sprintf("Preauthorized: %t\n", key.Preauthorized))
				if len(key.Tags) > 0 {
					result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(key.Tags, ", ")))
				}
				result.WriteString("\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Delete auth key tool
	server.AddTool(
		&mcp.Tool{
			Name:        "delete_auth_key",
			Description: "Delete an authentication key",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"key_id": {
						Type:        "string",
						Description: "ID of the authentication key to delete",
					},
				},
				Required: []string{"key_id"},
			},
		},
		mcp.ToolHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if api == nil || !api.IsAvailable() {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "API client not configured. Please set TAILSCALE_API_KEY environment variable."},
					},
				}, nil
			}

			var params struct {
				KeyID string `json:"key_id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if err := api.DeleteAuthKey(params.KeyID); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error deleting auth key: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Authentication key %s deleted successfully.", params.KeyID)},
				},
			}, nil
		}),
	)
}