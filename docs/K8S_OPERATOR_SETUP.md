# Tailscale Kubernetes Operator Setup Guide

## Quick Setup (5 Minutes)

### Step 1: Update ACL Policy

Go to https://login.tailscale.com/admin/acls and add these tags to your `tagOwners` section:

```json
"tagOwners": {
    "tag:k8s-operator": [],
    "tag:k8s": ["tag:k8s-operator"],
    // ... your existing tags ...
}
```

**What this does:**
- `tag:k8s-operator`: The operator device will have this tag
- `tag:k8s`: All devices created by the operator will have this tag

### Step 2: Create OAuth Client

1. Go to https://login.tailscale.com/admin/settings/oauth
2. Click **"Generate OAuth client"**
3. Configure:
   - **Description**: "Kubernetes Operator"
   - **Scopes**:
     - ✅ devices (Read/Write)
     - ✅ auth_keys (Write)
   - **Tags**: `tag:k8s-operator` (MUST match ACL)
4. Click **"Generate client"**
5. Save the credentials:
   - Client ID: `k123456CNTRL`
   - Client Secret: `tskey-client-k123456CNTRL-xxxx`

### Step 3: Enable K8s Support in MCP

```bash
# Add the MCP server with K8s support
claude mcp add -s user tailscale /path/to/tailscale-mcp \
  -e TAILSCALE_API_KEY=tskey-api-... \
  -e TAILSCALE_TAILNET=your-email@example.com \
  -e ENABLE_K8S_OPERATOR=true
```

### Step 4: Install the Operator

In Claude, run:
```
Install the Tailscale Kubernetes operator with:
- OAuth Client ID: k123456CNTRL
- OAuth Client Secret: tskey-client-k123456CNTRL-xxxx
```

## Why OAuth Instead of API Token?

| Component | Credential Type | Purpose | Location |
|-----------|----------------|---------|----------|
| MCP Server | API Token | Manage tailnet from your machine | Your computer |
| K8s Operator | OAuth Client | Create/manage devices from cluster | Inside Kubernetes |

**Security Benefits:**
- API token stays on your machine, not in the cluster
- OAuth client has limited, specific permissions
- Each component has its own identity for auditing
- Follows Tailscale's security best practices

## Troubleshooting

### "Tag not found" Error
- Ensure `tag:k8s-operator` exists in your ACL `tagOwners`
- Save the ACL policy before creating the OAuth client

### "Permission denied" Error
- Check OAuth client has `devices:write` scope
- Verify OAuth client is tagged with `tag:k8s-operator`

### ACL Validation
Run in Claude: `mcp__tailscale__k8s_prepare_acl` to see the exact configuration needed.

## Advanced Configuration

### Custom Tags for Different Environments

```json
"tagOwners": {
    "tag:k8s-operator": [],
    "tag:k8s": ["tag:k8s-operator"],
    "tag:k8s-prod": ["tag:k8s-operator"],
    "tag:k8s-staging": ["tag:k8s-operator"],
}
```

### Restrict K8s Device Access

Instead of allowing all connections, limit what K8s devices can access:

```json
"acls": [
    // K8s devices can only talk to each other
    {"action": "accept", "src": ["tag:k8s"], "dst": ["tag:k8s:*"]},
    // Users can access K8s services
    {"action": "accept", "src": ["autogroup:member"], "dst": ["tag:k8s:443", "tag:k8s:80"]},
    // ... other rules ...
]
```

## Complete Example ACL

```json
{
    "tagOwners": {
        "tag:k8s-operator": [],
        "tag:k8s": ["tag:k8s-operator"],
    },
    "acls": [
        {"action": "accept", "src": ["*"], "dst": ["*:*"]},
    ],
    "ssh": [
        {
            "action": "check",
            "src": ["autogroup:member"],
            "dst": ["autogroup:self"],
            "users": ["autogroup:nonroot", "root"],
        },
    ],
    "nodeAttrs": [
        {
            "target": ["autogroup:member"],
            "attr": ["funnel"],
        },
    ],
}
```

## Next Steps

After installation, you can:
- Create ProxyGroups for high availability
- Set up Ingress to expose services to your tailnet
- Configure Egress to access tailnet services from K8s
- Deploy Connectors for subnet routing

## Quick Commands Reference

```bash
# Check operator status
mcp__tailscale__k8s_operator_status

# Create an ingress
mcp__tailscale__k8s_ingress_create

# Create a ProxyGroup
mcp__tailscale__k8s_proxy_group_create

# List ProxyClasses
mcp__tailscale__k8s_proxy_class_list
```