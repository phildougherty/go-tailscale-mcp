package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Check environment variable for Kubernetes operator support
	enableK8sOperator := false
	if k8sEnv := os.Getenv("ENABLE_K8S_OPERATOR"); k8sEnv != "" {
		k8sEnv = strings.ToLower(k8sEnv)
		if k8sEnv == "true" || k8sEnv == "1" || k8sEnv == "yes" || k8sEnv == "on" {
			enableK8sOperator = true
		}
	}

	// Create and configure the MCP server
	srv, err := server.NewTailscaleServer(enableK8sOperator)
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