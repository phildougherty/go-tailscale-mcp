package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes client and provides Tailscale-specific operations
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient creates a new Kubernetes client with kubeconfig detection
func NewClient() (*Client, error) {
	config, err := getKubernetesConfig()
	if err != nil {
		return nil, NewKubeConfigError("failed to load Kubernetes configuration", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, NewConnectivityError("failed to create Kubernetes client", err)
	}

	// Test connectivity
	if err := testClusterConnectivity(clientset); err != nil {
		return nil, NewConnectivityError("failed to connect to Kubernetes cluster", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// getKubernetesConfig attempts to load Kubernetes configuration in order of preference:
// 1. In-cluster config (if running inside a pod)
// 2. KUBECONFIG environment variable
// 3. Default kubeconfig location (~/.kube/config)
func getKubernetesConfig() (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Try KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from KUBECONFIG=%s: %w", kubeconfig, err)
		}
		return config, nil
	}

	// Try default kubeconfig location
	if home := homedir.HomeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err == nil {
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", kubeconfig, err)
			}
			return config, nil
		}
	}

	return nil, fmt.Errorf("no valid Kubernetes configuration found. Please ensure:\n" +
		"1. KUBECONFIG environment variable is set, or\n" +
		"2. ~/.kube/config exists, or\n" +
		"3. Running inside a Kubernetes pod with service account")
}

// testClusterConnectivity tests basic connectivity to the Kubernetes cluster
func testClusterConnectivity(clientset *kubernetes.Clientset) error {
	_, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to get server version: %w", err)
	}
	return nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetConfig returns the Kubernetes rest config
func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// GetServerVersion returns the Kubernetes server version
func (c *Client) GetServerVersion() (string, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", NewConnectivityError("failed to get server version", err)
	}
	return version.String(), nil
}

// CheckPermissions checks if we have the necessary permissions for Tailscale operator operations
func (c *Client) CheckPermissions(ctx context.Context) error {
	// Check if we can access the tailscale-system namespace
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, "tailscale-system", metav1.GetOptions{})
	if err != nil {
		// Namespace might not exist yet, that's OK
		// But we should be able to create it
		return nil
	}

	// TODO: Add more specific permission checks for:
	// - Deployments, Services, Secrets in tailscale-system namespace
	// - ClusterRoles and ClusterRoleBindings
	// - Custom Resource Definitions
	// - ServiceAccounts

	return nil
}

// GetCurrentContext returns the current Kubernetes context
func (c *Client) GetCurrentContext() (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return "", NewKubeConfigError("failed to load raw config", err)
	}

	return rawConfig.CurrentContext, nil
}

// GetCurrentNamespace returns the current namespace
func (c *Client) GetCurrentNamespace() (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return "", NewKubeConfigError("failed to get current namespace", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return namespace, nil
}