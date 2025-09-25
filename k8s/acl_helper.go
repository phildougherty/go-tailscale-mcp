package k8s

import (
	"context"
	"encoding/json"
	"fmt"
)

// ACLConfig represents a Tailscale ACL configuration
type ACLConfig struct {
	Groups     map[string][]string    `json:"groups,omitempty"`
	TagOwners  map[string][]string    `json:"tagOwners,omitempty"`
	ACLs       []ACLRule              `json:"acls"`
	SSH        []SSHRule              `json:"ssh,omitempty"`
	NodeAttrs  []NodeAttr             `json:"nodeAttrs,omitempty"`
	Tests      []ACLTest              `json:"tests,omitempty"`
}

type ACLRule struct {
	Action     string   `json:"action"`
	Src        []string `json:"src"`
	Dst        []string `json:"dst"`
	SrcPosture []string `json:"srcPosture,omitempty"`
}

type SSHRule struct {
	Action string   `json:"action"`
	Src    []string `json:"src"`
	Dst    []string `json:"dst"`
	Users  []string `json:"users"`
}

type NodeAttr struct {
	Target []string `json:"target"`
	Attr   []string `json:"attr"`
}

type ACLTest struct {
	Src    string   `json:"src"`
	Accept []string `json:"accept,omitempty"`
	Deny   []string `json:"deny,omitempty"`
}

// PrepareK8sOperatorACL prepares or updates ACL configuration for Kubernetes operator
func PrepareK8sOperatorACL(currentACL string, operatorTag string) (string, error) {
	if operatorTag == "" {
		operatorTag = "tag:k8s-operator"
	}

	// Parse current ACL
	var aclConfig ACLConfig
	if currentACL != "" {
		if err := json.Unmarshal([]byte(currentACL), &aclConfig); err != nil {
			return "", fmt.Errorf("failed to parse current ACL: %w", err)
		}
	} else {
		// Start with a basic ACL structure if none exists
		aclConfig = ACLConfig{
			TagOwners: make(map[string][]string),
			ACLs: []ACLRule{
				{
					Action: "accept",
					Src:    []string{"*"},
					Dst:    []string{"*:*"},
				},
			},
		}
	}

	// Ensure tagOwners exists
	if aclConfig.TagOwners == nil {
		aclConfig.TagOwners = make(map[string][]string)
	}

	// Add required tags for Kubernetes operator
	requiredTags := map[string][]string{
		operatorTag: {},  // Empty means no owners (operator owns itself)
		"tag:k8s":   {operatorTag}, // operator can create devices with tag:k8s
	}

	// Merge required tags with existing ones
	for tag, owners := range requiredTags {
		if _, exists := aclConfig.TagOwners[tag]; !exists {
			aclConfig.TagOwners[tag] = owners
		} else if tag == "tag:k8s" {
			// Ensure operator is an owner of tag:k8s
			hasOperator := false
			for _, owner := range aclConfig.TagOwners[tag] {
				if owner == operatorTag {
					hasOperator = true
					break
				}
			}
			if !hasOperator {
				aclConfig.TagOwners[tag] = append(aclConfig.TagOwners[tag], operatorTag)
			}
		}
	}

	// Add SSH rule if not present (preserve existing SSH configuration)
	if len(aclConfig.SSH) == 0 {
		aclConfig.SSH = []SSHRule{
			{
				Action: "check",
				Src:    []string{"autogroup:member"},
				Dst:    []string{"autogroup:self"},
				Users:  []string{"autogroup:nonroot", "root"},
			},
		}
	}

	// Add node attributes if not present
	if len(aclConfig.NodeAttrs) == 0 {
		aclConfig.NodeAttrs = []NodeAttr{
			{
				Target: []string{"autogroup:member"},
				Attr:   []string{"funnel"},
			},
		}
	}

	// Marshal back to JSON with proper formatting
	jsonBytes, err := json.MarshalIndent(aclConfig, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ACL: %w", err)
	}

	return string(jsonBytes), nil
}

// GenerateK8sOperatorACLInstructions generates user-friendly instructions
func GenerateK8sOperatorACLInstructions() string {
	return `
=== Kubernetes Operator ACL Setup Instructions ===

The Tailscale Kubernetes Operator requires specific tags to be configured in your ACL policy.

REQUIRED TAGS:
1. tag:k8s-operator - The operator itself will be tagged with this
2. tag:k8s - Devices created by the operator will use this tag

MANUAL STEPS:
1. Go to: https://login.tailscale.com/admin/acls
2. Add the following to your "tagOwners" section:

    "tagOwners": {
        "tag:k8s-operator": [],
        "tag:k8s": ["tag:k8s-operator"],
        // ... your existing tags ...
    }

3. Save the ACL policy

WHAT THIS MEANS:
- tag:k8s-operator: [] means the operator owns itself (no other owners)
- tag:k8s: ["tag:k8s-operator"] means the operator can create devices with tag:k8s

OAUTH CLIENT SETUP:
When creating the OAuth client, assign it the tag "tag:k8s-operator"

OPTIONAL ADDITIONAL TAGS:
If you want to use custom tags for different purposes:
- tag:k8s-ingress - For ingress proxies
- tag:k8s-egress - For egress proxies
- tag:k8s-connector - For subnet routers/exit nodes

Make sure tag:k8s-operator is an owner of any custom tags you want to use.
`
}

// ValidateOperatorTags checks if the ACL has the required tags configured
func ValidateOperatorTags(aclJSON string) (bool, []string) {
	var aclConfig ACLConfig
	if err := json.Unmarshal([]byte(aclJSON), &aclConfig); err != nil {
		return false, []string{"Failed to parse ACL JSON"}
	}

	issues := []string{}

	// Check if tag:k8s-operator exists
	if _, exists := aclConfig.TagOwners["tag:k8s-operator"]; !exists {
		issues = append(issues, "Missing tag:k8s-operator in tagOwners")
	}

	// Check if tag:k8s exists and has k8s-operator as owner
	if owners, exists := aclConfig.TagOwners["tag:k8s"]; !exists {
		issues = append(issues, "Missing tag:k8s in tagOwners")
	} else {
		hasOperator := false
		for _, owner := range owners {
			if owner == "tag:k8s-operator" {
				hasOperator = true
				break
			}
		}
		if !hasOperator {
			issues = append(issues, "tag:k8s-operator is not an owner of tag:k8s")
		}
	}

	return len(issues) == 0, issues
}

// SetupOperatorACL is a high-level function that sets up ACLs for the operator
func (c *Client) SetupOperatorACL(ctx context.Context, apiClient interface{}) error {
	// This would integrate with your API client to actually update the ACLs
	// For now, it returns instructions
	fmt.Println(GenerateK8sOperatorACLInstructions())
	return nil
}