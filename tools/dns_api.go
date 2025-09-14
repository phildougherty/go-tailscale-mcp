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

// RegisterDNSAPITools registers DNS management tools using the API
func RegisterDNSAPITools(server *mcp.Server, api *tailscale.APIClient) {
	// Get DNS configuration tool
	server.AddTool(
		&mcp.Tool{
			Name:        "get_dns_config",
			Description: "Get the current DNS configuration",
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

			dnsConfig, err := api.GetDNS()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error getting DNS configuration: %v", err)},
					},
				}, nil
			}

			var result strings.Builder
			result.WriteString("DNS Configuration:\n\n")
			result.WriteString(fmt.Sprintf("MagicDNS Enabled: %t\n", dnsConfig.MagicDNS))

			if len(dnsConfig.Nameservers) > 0 {
				result.WriteString(fmt.Sprintf("Nameservers:\n"))
				for _, ns := range dnsConfig.Nameservers {
					result.WriteString(fmt.Sprintf("  - %s\n", ns))
				}
			} else {
				result.WriteString("Nameservers: None configured\n")
			}

			if len(dnsConfig.Domains) > 0 {
				result.WriteString(fmt.Sprintf("Search Domains:\n"))
				for _, domain := range dnsConfig.Domains {
					result.WriteString(fmt.Sprintf("  - %s\n", domain))
				}
			} else {
				result.WriteString("Search Domains: None configured\n")
			}

			if len(dnsConfig.Routes) > 0 {
				result.WriteString("DNS Routes:\n")
				for domain, servers := range dnsConfig.Routes {
					result.WriteString(fmt.Sprintf("  %s -> %s\n", domain, strings.Join(servers, ", ")))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result.String()},
				},
			}, nil
		}),
	)

	// Set DNS nameservers tool
	server.AddTool(
		&mcp.Tool{
			Name:        "set_dns_nameservers",
			Description: "Set DNS nameservers",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"nameservers": {
						Type: "array",
						Items: &jsonschema.Schema{Type: "string"},
						Description: "List of DNS nameserver IP addresses (e.g., ['8.8.8.8', '1.1.1.1'])",
					},
				},
				Required: []string{"nameservers"},
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
				Nameservers []string `json:"nameservers"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if len(params.Nameservers) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "No nameservers specified. Please provide at least one nameserver."},
					},
				}, nil
			}

			if err := api.SetDNSNameservers(params.Nameservers); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error setting DNS nameservers: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("DNS nameservers updated successfully: %s", strings.Join(params.Nameservers, ", "))},
				},
			}, nil
		}),
	)

	// Set DNS preferences tool (MagicDNS)
	server.AddTool(
		&mcp.Tool{
			Name:        "set_dns_preferences",
			Description: "Set DNS preferences including MagicDNS on/off",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"magic_dns": {
						Type:        "boolean",
						Description: "Whether to enable MagicDNS",
					},
				},
				Required: []string{"magic_dns"},
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
				MagicDNS bool `json:"magic_dns"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if err := api.SetDNSPreferences(params.MagicDNS); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error setting DNS preferences: %v", err)},
					},
				}, nil
			}

			status := "disabled"
			if params.MagicDNS {
				status = "enabled"
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("MagicDNS %s successfully.", status)},
				},
			}, nil
		}),
	)

	// Set DNS search paths tool
	server.AddTool(
		&mcp.Tool{
			Name:        "set_dns_search_paths",
			Description: "Set DNS search paths",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"search_paths": {
						Type: "array",
						Items: &jsonschema.Schema{Type: "string"},
						Description: "List of DNS search domain paths (e.g., ['example.com', 'company.local'])",
					},
				},
				Required: []string{"search_paths"},
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
				SearchPaths []string `json:"search_paths"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
					},
				}, nil
			}

			if len(params.SearchPaths) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "No search paths specified. Please provide at least one search path."},
					},
				}, nil
			}

			if err := api.SetDNSSearchPaths(params.SearchPaths); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error setting DNS search paths: %v", err)},
					},
				}, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("DNS search paths updated successfully: %s", strings.Join(params.SearchPaths, ", "))},
				},
			}, nil
		}),
	)
}