package server

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/tailscale"
	"github.com/phildougherty/go-tailscale-mcp/tools"
)

type TailscaleServer struct {
	*mcp.Server
	cli *tailscale.CLI
}

func NewTailscaleServer() (*TailscaleServer, error) {
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

	ts := &TailscaleServer{
		Server: server,
		cli:    cli,
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
	tools.RegisterDeviceTools(s.Server, s.cli)
	tools.RegisterNetworkTools(s.Server, s.cli)
	tools.RegisterRoutingTools(s.Server, s.cli)
	tools.RegisterSystemTools(s.Server, s.cli)

	return nil
}