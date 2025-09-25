package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TailscaleSystemNamespace = "tailscale"
	OperatorDeploymentName   = "operator"
	OperatorServiceAccount   = "operator"
	DefaultOperatorImage     = "tailscale/k8s-operator:latest"
)

// OperatorStatus represents the status of the Tailscale operator
type OperatorStatus struct {
	Installed         bool                `json:"installed"`
	Healthy           bool                `json:"healthy"`
	Version           string              `json:"version,omitempty"`
	Replicas          int32               `json:"replicas"`
	ReadyReplicas     int32               `json:"ready_replicas"`
	Namespace         string              `json:"namespace"`
	LastUpdateTime    *metav1.Time        `json:"last_update_time,omitempty"`
	Conditions        []appsv1.DeploymentCondition `json:"conditions,omitempty"`
	ErrorMessage      string              `json:"error_message,omitempty"`
}

// Installation and upgrade functions removed - use official Tailscale installation methods:
// https://tailscale.com/kb/1236/kubernetes-operator

// GetOperatorStatus returns the current status of the Tailscale operator
func (c *Client) GetOperatorStatus(ctx context.Context) (*OperatorStatus, error) {
	status := &OperatorStatus{
		Namespace: TailscaleSystemNamespace,
	}

	// Check if namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, TailscaleSystemNamespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.ErrorMessage = "tailscale namespace not found"
			return status, nil
		}
		return nil, NewConnectivityError("failed to check namespace", err)
	}

	// Check deployment
	deployment, err := c.clientset.AppsV1().Deployments(TailscaleSystemNamespace).Get(ctx, OperatorDeploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.ErrorMessage = "operator deployment not found"
			return status, nil
		}
		return nil, NewResourceNotFoundError("deployment", OperatorDeploymentName, err)
	}

	status.Installed = true
	status.Replicas = *deployment.Spec.Replicas
	status.ReadyReplicas = deployment.Status.ReadyReplicas
	status.Healthy = deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == *deployment.Spec.Replicas
	status.Conditions = deployment.Status.Conditions

	// Get version from image
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		status.Version = deployment.Spec.Template.Spec.Containers[0].Image
	}

	// Check last update time
	if deployment.Status.ObservedGeneration > 0 {
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == appsv1.DeploymentProgressing {
				status.LastUpdateTime = &condition.LastUpdateTime
				break
			}
		}
	}

	return status, nil
}

// InstallOperator has been removed - use official installation methods:
// kubectl apply -f https://tailscale.com/install/kubernetes/operator.yaml
// or Helm: helm install tailscale-operator tailscale/tailscale-operator

// UpgradeOperator has been removed - use kubectl or helm to upgrade

// Helper functions for namespace, secrets and service accounts removed
// These are handled by the official operator installation

// Deployment creation and wait functions removed - handled by official installation