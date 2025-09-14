package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/tailscale"
)

// RegisterACLTools registers ACL management tools
func RegisterACLTools(server *mcp.Server, api *tailscale.APIClient) {
	// Get ACL tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_acl",
			Description: "Get the current ACL (Access Control List) policy",
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

			acl, err := api.GetACL()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting ACL: %v", err)},
					},
				}, nil
			}

			// Return the raw HuJSON policy
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Current ACL Policy (HuJSON format):\n\n%s", acl.RawPolicy)},
				},
			}, nil
		}),
	)

	// Update ACL tool
	server.AddTool(
		&mcp.Tool{
			Name:        "update_acl",
			Description: "Update the ACL (Access Control List) policy",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"acl": {
						Type:        "string",
						Description: "ACL policy in JSON format",
					},
				},
				Required: []string{"acl"},
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
				ACL string `json:"acl"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Try to parse as JSON first, otherwise treat as HuJSON
			var acl tailscale.ACL
			if err := json.Unmarshal([]byte(params.ACL), &acl); err != nil {
				// Not valid JSON, treat as HuJSON
				acl.RawPolicy = params.ACL
			}

			// Validate the ACL first
			if err := api.ValidateACL(&acl); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("ACL validation failed: %v", err)},
					},
				}, nil
			}

			// Update the ACL
			if err := api.SetACL(&acl); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error updating ACL: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "ACL policy updated successfully."},
				},
			}, nil
		}),
	)

	// Validate ACL tool
	server.AddTool(
		&mcp.Tool{
			Name:        "validate_acl",
			Description: "Validate an ACL (Access Control List) policy without applying it",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"acl": {
						Type:        "string",
						Description: "ACL policy in JSON format to validate",
					},
				},
				Required: []string{"acl"},
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
				ACL string `json:"acl"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			// Try to parse as JSON first, otherwise treat as HuJSON
			var acl tailscale.ACL
			if err := json.Unmarshal([]byte(params.ACL), &acl); err != nil {
				// Not valid JSON, treat as HuJSON
				acl.RawPolicy = params.ACL
			}

			// Validate the ACL
			if err := api.ValidateACL(&acl); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("ACL validation failed: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "ACL policy is valid."},
				},
			}, nil
		}),
	)
}