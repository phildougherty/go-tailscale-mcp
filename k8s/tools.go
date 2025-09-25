package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegisterK8sOperatorTools registers all Kubernetes operator tools with the MCP server
func RegisterK8sOperatorTools(server *mcp.Server) error {
	// ACL preparation tool
	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_prepare_acl",
			Description: "Prepare Tailscale ACL configuration for Kubernetes operator (shows required configuration)",
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
		},
		mcp.ToolHandler(handlePrepareACL),
	)

	// Operator management tools
	// Operator installation removed - install manually using kubectl or helm
	// The operator requires proper RBAC, CRDs, and configuration that are
	// better handled through official Tailscale installation methods:
	// https://tailscale.com/kb/1236/kubernetes-operator

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_operator_status",
			Description: "Get the status of the Tailscale Kubernetes operator",
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
			},
		},
		mcp.ToolHandler(handleOperatorStatus),
	)

	// ProxyClass management
	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_class_create",
			Description: "Create a ProxyClass resource for customizing proxy configurations",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":        {Type: "string", Description: "Name of the ProxyClass"},
					"namespace":   {Type: "string", Description: "Namespace for the ProxyClass"},
					"labels":      {Type: "object", Description: "Labels to apply to proxy pods"},
					"annotations": {Type: "object", Description: "Annotations to apply to proxy pods"},
				},
				Required: []string{"name", "namespace"},
			},
		},
		mcp.ToolHandler(handleProxyClassCreate),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_class_list",
			Description: "List ProxyClass resources in a namespace",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespace": {Type: "string", Description: "Namespace to list ProxyClasses from (empty for all)"},
				},
			},
		},
		mcp.ToolHandler(handleProxyClassList),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_class_delete",
			Description: "Delete a ProxyClass resource",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":      {Type: "string", Description: "Name of the ProxyClass to delete"},
					"namespace": {Type: "string", Description: "Namespace of the ProxyClass"},
				},
				Required: []string{"name", "namespace"},
			},
		},
		mcp.ToolHandler(handleProxyClassDelete),
	)

	// ProxyGroup management
	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_group_create",
			Description: "Create a ProxyGroup for high availability configurations",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":        {Type: "string", Description: "Name of the ProxyGroup"},
					"namespace":   {Type: "string", Description: "Namespace for the ProxyGroup"},
					"type":        {Type: "string", Description: "Type of ProxyGroup (egress or ingress)"},
					"replicas":    {Type: "integer", Description: "Number of replicas"},
					"proxy_class": {Type: "string", Description: "ProxyClass to use for configuration"},
					"tags": {
						Type:        "array",
						Items:       &jsonschema.Schema{Type: "string"},
						Description: "Tags to apply to the proxy devices",
					},
				},
				Required: []string{"name", "namespace", "type"},
			},
		},
		mcp.ToolHandler(handleProxyGroupCreate),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_group_status",
			Description: "Get the status of a ProxyGroup",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":      {Type: "string", Description: "Name of the ProxyGroup"},
					"namespace": {Type: "string", Description: "Namespace of the ProxyGroup"},
				},
				Required: []string{"name", "namespace"},
			},
		},
		mcp.ToolHandler(handleProxyGroupStatus),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_proxy_group_scale",
			Description: "Scale a ProxyGroup to a different number of replicas",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":      {Type: "string", Description: "Name of the ProxyGroup"},
					"namespace": {Type: "string", Description: "Namespace of the ProxyGroup"},
					"replicas":  {Type: "integer", Description: "New number of replicas"},
				},
				Required: []string{"name", "namespace", "replicas"},
			},
		},
		mcp.ToolHandler(handleProxyGroupScale),
	)

	// Ingress and Egress
	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_ingress_create",
			Description: "Create a Tailscale ingress to expose a cluster service to the tailnet",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":         {Type: "string", Description: "Name of the ingress"},
					"namespace":    {Type: "string", Description: "Namespace for the ingress"},
					"hostname":     {Type: "string", Description: "Hostname for the ingress"},
					"service_name": {Type: "string", Description: "Name of the service to expose"},
					"service_port": {Type: "integer", Description: "Port of the service to expose"},
				},
				Required: []string{"name", "namespace", "hostname", "service_name", "service_port"},
			},
		},
		mcp.ToolHandler(handleIngressCreate),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_egress_create",
			Description: "Create an egress service to expose a tailnet service to the cluster",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":              {Type: "string", Description: "Name of the egress service"},
					"namespace":         {Type: "string", Description: "Namespace for the egress service"},
					"external_hostname": {Type: "string", Description: "External hostname to connect to"},
					"port":              {Type: "integer", Description: "Port to connect to"},
				},
				Required: []string{"name", "namespace", "external_hostname", "port"},
			},
		},
		mcp.ToolHandler(handleEgressCreate),
	)

	// Connector and DNSConfig
	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_connector_create",
			Description: "Create a Connector for subnet routing or exit node functionality",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":        {Type: "string", Description: "Name of the Connector"},
					"namespace":   {Type: "string", Description: "Namespace for the Connector"},
					"hostname":    {Type: "string", Description: "Hostname for the Connector"},
					"proxy_class": {Type: "string", Description: "ProxyClass to use"},
					"subnet_routes": {
						Type:        "array",
						Items:       &jsonschema.Schema{Type: "string"},
						Description: "Subnet routes to advertise",
					},
					"exit_node": {Type: "boolean", Description: "Enable exit node functionality"},
					"tags": {
						Type:        "array",
						Items:       &jsonschema.Schema{Type: "string"},
						Description: "Tags to apply to the Connector",
					},
				},
				Required: []string{"name", "namespace"},
			},
		},
		mcp.ToolHandler(handleConnectorCreate),
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "mcp__tailscale__k8s_dns_config_create",
			Description: "Create a DNSConfig for MagicDNS configuration",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":      {Type: "string", Description: "Name of the DNSConfig"},
					"namespace": {Type: "string", Description: "Namespace for the DNSConfig"},
					"magic_dns": {Type: "boolean", Description: "Enable MagicDNS"},
					"nameservers": {
						Type:        "array",
						Items:       &jsonschema.Schema{Type: "string"},
						Description: "List of nameserver IPs",
					},
				},
				Required: []string{"name", "namespace", "magic_dns"},
			},
		},
		mcp.ToolHandler(handleDNSConfigCreate),
	)

	return nil
}

// Tool handlers

func handlePrepareACL(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instructions := GenerateK8sOperatorACLInstructions()

	// Also provide a sample ACL configuration
	sampleACL := `
=== SAMPLE ACL CONFIGURATION ===

Copy this into your Tailscale ACL editor at https://login.tailscale.com/admin/acls

{
    "tagOwners": {
        "tag:k8s-operator": [],
        "tag:k8s": ["tag:k8s-operator"],
        // Add any custom tags here:
        // "tag:k8s-ingress": ["tag:k8s-operator"],
        // "tag:k8s-egress": ["tag:k8s-operator"],
    },
    "acls": [
        // Your existing ACL rules...
        {"action": "accept", "src": ["*"], "dst": ["*:*"]},

        // Optional: Add specific rules for k8s devices
        // {"action": "accept", "src": ["tag:k8s"], "dst": ["tag:k8s:*"]},
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

=== OAUTH CLIENT CONFIGURATION ===

When creating the OAuth client at https://login.tailscale.com/admin/settings/oauth

1. Click "Generate OAuth client"
2. Set the description (e.g., "Kubernetes Operator")
3. Select scopes:
   - devices:write (Create and manage devices)
   - auth_keys:write (Create auth keys)
   - routes:write (optional, for subnet routing)
   - dns:write (optional, for MagicDNS)

4. IMPORTANT: Add tags: tag:k8s-operator
   (This must match the tag in your ACL policy)

5. Click "Generate client"
6. Copy the client ID and secret

The OAuth client will look like:
- Client ID: k123456CNTRL
- Client Secret: tskey-client-k123456CNTRL-xxxxxxxxxxxx
`

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: instructions + sampleACL},
		},
	}, nil
}

// Removed handleOperatorInstall - operator should be installed using official methods

func handleOperatorStatus(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := NewClient()
	if err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	status, err := client.GetOperatorStatus(ctx)
	if err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Operator Status:\n%s", string(statusJSON))},
		},
	}, nil
}

// Removed handleOperatorUpgrade - operator should be upgraded using official methods

func handleProxyClassCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name        string                 `json:"name"`
		Namespace   string                 `json:"namespace"`
		Labels      map[string]interface{} `json:"labels,omitempty"`
		Annotations map[string]interface{} `json:"annotations,omitempty"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	proxyClass := &ProxyClass{
		Metadata: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
		},
		Spec: ProxyClassSpec{},
	}

	// Add labels if provided
	if params.Labels != nil {
		labelsStr := make(map[string]string)
		for k, v := range params.Labels {
			labelsStr[k] = fmt.Sprintf("%v", v)
		}
		if proxyClass.Spec.StatefulSet == nil {
			proxyClass.Spec.StatefulSet = &StatefulSetSpec{}
		}
		if proxyClass.Spec.StatefulSet.Pod == nil {
			proxyClass.Spec.StatefulSet.Pod = &PodSpec{}
		}
		proxyClass.Spec.StatefulSet.Pod.Labels = labelsStr
	}

	// Add annotations if provided
	if params.Annotations != nil {
		annotationsStr := make(map[string]string)
		for k, v := range params.Annotations {
			annotationsStr[k] = fmt.Sprintf("%v", v)
		}
		if proxyClass.Spec.StatefulSet == nil {
			proxyClass.Spec.StatefulSet = &StatefulSetSpec{}
		}
		if proxyClass.Spec.StatefulSet.Pod == nil {
			proxyClass.Spec.StatefulSet.Pod = &PodSpec{}
		}
		proxyClass.Spec.StatefulSet.Pod.Annotations = annotationsStr
	}

	if err := rm.CreateProxyClass(ctx, proxyClass); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyClass '%s' created successfully in namespace '%s'",
				proxyClass.Metadata.Name, proxyClass.Metadata.Namespace)},
		},
	}, nil
}

func handleProxyClassList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Namespace string `json:"namespace,omitempty"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	proxyClasses, err := rm.ListProxyClasses(ctx, params.Namespace)
	if err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	listJSON, err := json.MarshalIndent(proxyClasses, "", "  ")
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyClasses:\n%s", string(listJSON))},
		},
	}, nil
}

func handleProxyClassDelete(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	if err := rm.DeleteProxyClass(ctx, params.Namespace, params.Name); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyClass '%s' deleted from namespace '%s'", params.Name, params.Namespace)},
		},
	}, nil
}

func handleProxyGroupCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name       string   `json:"name"`
		Namespace  string   `json:"namespace"`
		Type       string   `json:"type"`
		Replicas   int32    `json:"replicas,omitempty"`
		ProxyClass string   `json:"proxy_class,omitempty"`
		Tags       []string `json:"tags,omitempty"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	replicas := params.Replicas
	if replicas == 0 {
		replicas = 2 // Default
	}

	proxyGroup := &ProxyGroup{
		Metadata: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
		},
		Spec: ProxyGroupSpec{
			Type:       params.Type,
			Replicas:   &replicas,
			ProxyClass: params.ProxyClass,
			Tags:       params.Tags,
		},
	}

	if err := rm.CreateProxyGroup(ctx, proxyGroup); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyGroup '%s' created successfully in namespace '%s' with %d replicas",
				proxyGroup.Metadata.Name, proxyGroup.Metadata.Namespace, replicas)},
		},
	}, nil
}

func handleProxyGroupStatus(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	status, err := rm.GetProxyGroupStatus(ctx, params.Namespace, params.Name)
	if err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyGroup Status:\n%s", string(statusJSON))},
		},
	}, nil
}

func handleProxyGroupScale(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Replicas  int32  `json:"replicas"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	if err := rm.ScaleProxyGroup(ctx, params.Namespace, params.Name, params.Replicas); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("ProxyGroup '%s' scaled to %d replicas", params.Name, params.Replicas)},
		},
	}, nil
}

func handleIngressCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Hostname    string `json:"hostname"`
		ServiceName string `json:"service_name"`
		ServicePort int32  `json:"service_port"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	if err := rm.CreateTailscaleIngress(ctx, params.Namespace, params.Name, params.Hostname, params.ServiceName, params.ServicePort); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Tailscale ingress '%s' created successfully. Service '%s:%d' will be exposed as '%s'",
				params.Name, params.ServiceName, params.ServicePort, params.Hostname)},
		},
	}, nil
}

func handleEgressCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name             string `json:"name"`
		Namespace        string `json:"namespace"`
		ExternalHostname string `json:"external_hostname"`
		Port             int32  `json:"port"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	if err := rm.CreateEgressService(ctx, params.Namespace, params.Name, params.ExternalHostname, params.Port); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Egress service '%s' created successfully. External service '%s:%d' is now accessible in the cluster",
				params.Name, params.ExternalHostname, params.Port)},
		},
	}, nil
}

func handleConnectorCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name         string   `json:"name"`
		Namespace    string   `json:"namespace"`
		Hostname     string   `json:"hostname,omitempty"`
		ProxyClass   string   `json:"proxy_class,omitempty"`
		SubnetRoutes []string `json:"subnet_routes,omitempty"`
		ExitNode     bool     `json:"exit_node,omitempty"`
		Tags         []string `json:"tags,omitempty"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		Metadata: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
		},
		Spec: ConnectorSpec{
			Hostname:   params.Hostname,
			ProxyClass: params.ProxyClass,
			ExitNode:   params.ExitNode,
			Tags:       params.Tags,
		},
	}

	// Handle subnet routes
	if len(params.SubnetRoutes) > 0 {
		connector.Spec.SubnetRouter = &SubnetRouterSpec{
			AdvertiseRoutes: params.SubnetRoutes,
		}
	}

	if err := rm.CreateConnector(ctx, connector); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Connector '%s' created successfully in namespace '%s'",
				connector.Metadata.Name, connector.Metadata.Namespace)},
		},
	}, nil
}

func handleDNSConfigCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name        string   `json:"name"`
		Namespace   string   `json:"namespace"`
		MagicDNS    bool     `json:"magic_dns"`
		Nameservers []string `json:"nameservers,omitempty"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)},
			},
		}, nil
	}

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	rm, err := NewResourceManager(client)
	if err != nil {
		return nil, err
	}

	dnsConfig := &DNSConfig{
		Metadata: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
		},
		Spec: DNSConfigSpec{
			Nameserver: NameserverSpec{
				// The nameserver will use default image if not specified
			},
		},
	}

	if err := rm.CreateDNSConfig(ctx, dnsConfig); err != nil {
		if k8sErr, ok := err.(*K8sError); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: k8sErr.FormatErrorWithHint()},
				},
			}, nil
		}
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("DNSConfig '%s' created successfully in namespace '%s'",
				dnsConfig.Metadata.Name, dnsConfig.Metadata.Namespace)},
		},
	}, nil
}