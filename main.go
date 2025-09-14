package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phildougherty/go-tailscale-mcp/server"
)

func main() {
	// Simple version check
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("tailscale-mcp v1.0.0")
		return
	}

	ctx := context.Background()

	// Create and configure the MCP server
	srv, err := server.NewTailscaleServer()
	if err != nil {
		log.Fatalf("Failed to create Tailscale MCP server: %v", err)
	}

	// Run the server with stdio transport
	transport := &mcp.StdioTransport{}
	if err := srv.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}