#!/bin/bash

# Simple test script for the Tailscale MCP server

echo "Testing Tailscale MCP Server..."

# Test initialize request
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2025-01-14","clientInfo":{"name":"test","version":"1.0.0"}}}' | timeout 2 ./tailscale-mcp 2>/dev/null | grep -q "tailscale-mcp"

if [ $? -eq 0 ]; then
    echo "✓ Server responds to initialize request"
else
    echo "✗ Server failed to respond properly"
    exit 1
fi

echo "✓ Basic test passed!"