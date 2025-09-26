package server

import (
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/k8s"
	"github.com/phildougherty/go-tailscale-mcp/tailscale"
	"github.com/phildougherty/go-tailscale-mcp/tools"
)

type TailscaleServer struct {
	*mcp.Server
	cli              *tailscale.CLI
	api              *tailscale.APIClient
	enableK8sOperator bool
}

func NewTailscaleServer(enableK8sOperator bool) (*TailscaleServer, error) {
	// Initialize the MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "tailscale-mcp",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			HasTools: true,
		},
	)

	// Create Tailscale CLI wrapper
	cli := tailscale.NewCLI()

	// Create API client if API key is provided
	var apiClient *tailscale.APIClient
	if apiKey := os.Getenv("TAILSCALE_API_KEY"); apiKey != "" {
		var err error

		// Check if tailnet is explicitly provided
		if tailnet := os.Getenv("TAILSCALE_TAILNET"); tailnet != "" {
			apiClient, err = tailscale.NewAPIClientWithTailnet(apiKey, tailnet)
		} else {
			apiClient, err = tailscale.NewAPIClient(apiKey)
		}

		if err != nil {
			// Log error but continue without API
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize Tailscale API client: %v\n", err)
			fmt.Fprintf(os.Stderr, "Hint: Set TAILSCALE_TAILNET environment variable to your tailnet domain (e.g., your-email@example.com)\n")
		} else {
			fmt.Fprintf(os.Stderr, "Tailscale API client initialized successfully\n")
		}
	}

	ts := &TailscaleServer{
		Server:           server,
		cli:              cli,
		api:              apiClient,
		enableK8sOperator: enableK8sOperator,
	}

	// Register all tools
	if err := ts.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return ts, nil
}

func (s *TailscaleServer) registerTools() error {
	// Register all Tailscale tool categories
	tools.RegisterProfileTools(s.Server, s.cli)
	tools.RegisterDeviceToolsWithAPI(s.Server, s.cli, s.api)
	tools.RegisterNetworkTools(s.Server, s.cli)
	tools.RegisterRoutingToolsWithAPI(s.Server, s.cli, s.api)
	tools.RegisterSystemTools(s.Server, s.cli)
	tools.RegisterDiagnosticTools(s.Server, s.cli)

	// Register API-specific tools if API is available
	if s.api != nil && s.api.IsAvailable() {
		tools.RegisterACLTools(s.Server, s.api)
		tools.RegisterAuthKeyTools(s.Server, s.api)
		tools.RegisterDNSAPITools(s.Server, s.api)
	}

	// Register Kubernetes operator tools if enabled
	if s.enableK8sOperator {
		if err := k8s.RegisterK8sOperatorTools(s.Server); err != nil {
			return fmt.Errorf("failed to register Kubernetes operator tools: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Kubernetes operator tools enabled\n")
	}

	return nil
}