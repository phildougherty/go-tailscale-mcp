package k8s

import (
	"fmt"
)

// K8sError represents a Kubernetes-specific error
type K8sError struct {
	Type    ErrorType
	Message string
	Cause   error
}

type ErrorType string

const (
	// Configuration errors
	ErrorTypeKubeConfig    ErrorType = "kubeconfig"
	ErrorTypePermission    ErrorType = "permission"
	ErrorTypeConnectivity  ErrorType = "connectivity"

	// Resource errors
	ErrorTypeResourceNotFound ErrorType = "resource_not_found"
	ErrorTypeResourceConflict ErrorType = "resource_conflict"
	ErrorTypeResourceInvalid  ErrorType = "resource_invalid"

	// Operator errors
	ErrorTypeOperatorNotFound ErrorType = "operator_not_found"
	ErrorTypeOperatorInstall  ErrorType = "operator_install"
	ErrorTypeOperatorUpgrade  ErrorType = "operator_upgrade"

	// General errors
	ErrorTypeUnknown ErrorType = "unknown"
)

func (e *K8sError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

func (e *K8sError) Unwrap() error {
	return e.Cause
}

// NewK8sError creates a new Kubernetes error
func NewK8sError(errorType ErrorType, message string, cause error) *K8sError {
	return &K8sError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// Helper functions for common error types
func NewKubeConfigError(message string, cause error) *K8sError {
	return NewK8sError(ErrorTypeKubeConfig, message, cause)
}

func NewPermissionError(message string, cause error) *K8sError {
	return NewK8sError(ErrorTypePermission, message, cause)
}

func NewConnectivityError(message string, cause error) *K8sError {
	return NewK8sError(ErrorTypeConnectivity, message, cause)
}

func NewResourceNotFoundError(resource, name string, cause error) *K8sError {
	message := fmt.Sprintf("%s '%s' not found", resource, name)
	return NewK8sError(ErrorTypeResourceNotFound, message, cause)
}

func NewResourceConflictError(resource, name string, cause error) *K8sError {
	message := fmt.Sprintf("%s '%s' already exists", resource, name)
	return NewK8sError(ErrorTypeResourceConflict, message, cause)
}

func NewOperatorNotFoundError(cause error) *K8sError {
	return NewK8sError(ErrorTypeOperatorNotFound, "Tailscale operator is not installed", cause)
}

func NewOperatorInstallError(message string, cause error) *K8sError {
	return NewK8sError(ErrorTypeOperatorInstall, message, cause)
}

// GetTroubleshootingHint returns a helpful troubleshooting hint for the error
func (e *K8sError) GetTroubleshootingHint() string {
	switch e.Type {
	case ErrorTypeKubeConfig:
		return "Troubleshooting tips:\n" +
			"1. Ensure kubectl is installed and configured\n" +
			"2. Check if KUBECONFIG environment variable is set\n" +
			"3. Verify ~/.kube/config exists and is readable\n" +
			"4. Test connection with: kubectl cluster-info"

	case ErrorTypePermission:
		return "Troubleshooting tips:\n" +
			"1. Check if you have sufficient RBAC permissions\n" +
			"2. Verify service account permissions\n" +
			"3. Try: kubectl auth can-i '*' '*' --all-namespaces\n" +
			"4. Contact your cluster administrator"

	case ErrorTypeConnectivity:
		return "Troubleshooting tips:\n" +
			"1. Verify cluster connectivity: kubectl cluster-info\n" +
			"2. Check network connectivity to Kubernetes API server\n" +
			"3. Verify VPN/proxy settings if applicable\n" +
			"4. Check if cluster certificates are valid"

	case ErrorTypeOperatorNotFound:
		return "Troubleshooting tips:\n" +
			"1. Install the Tailscale operator: kubectl apply -f https://github.com/tailscale/tailscale/raw/main/cmd/k8s-operator/deploy/manifests/operator.yaml\n" +
			"2. Or use the tool: mcp__tailscale__k8s_operator_install\n" +
			"3. Check operator status: kubectl get pods -n tailscale\n" +
			"4. Verify operator deployment: kubectl get deployment -n tailscale"

	case ErrorTypeResourceConflict:
		return "Troubleshooting tips:\n" +
			"1. Check existing resource: kubectl get <resource-type> <name>\n" +
			"2. Delete if no longer needed: kubectl delete <resource-type> <name>\n" +
			"3. Use a different name for the resource\n" +
			"4. Update the existing resource instead of creating new"

	default:
		return "General troubleshooting tips:\n" +
			"1. Check kubectl configuration: kubectl config view\n" +
			"2. Verify cluster access: kubectl cluster-info\n" +
			"3. Check logs for more details\n" +
			"4. Consult Tailscale Kubernetes operator documentation"
	}
}

// FormatErrorWithHint formats the error with troubleshooting hints
func (e *K8sError) FormatErrorWithHint() string {
	return fmt.Sprintf("%s\n\n%s", e.Error(), e.GetTroubleshootingHint())
}