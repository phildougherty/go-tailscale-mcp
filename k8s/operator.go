package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	TailscaleSystemNamespace = "tailscale-system"
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

// InstallOperatorOptions represents options for installing the operator
type InstallOperatorOptions struct {
	OAuthClientID     string `json:"oauth_client_id"`
	OAuthClientSecret string `json:"oauth_client_secret"`
	Image             string `json:"image,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
}

// UpgradeOperatorOptions represents options for upgrading the operator
type UpgradeOperatorOptions struct {
	Image   string `json:"image"`
	Force   bool   `json:"force,omitempty"`
}

// GetOperatorStatus returns the current status of the Tailscale operator
func (c *Client) GetOperatorStatus(ctx context.Context) (*OperatorStatus, error) {
	status := &OperatorStatus{
		Namespace: TailscaleSystemNamespace,
	}

	// Check if namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, TailscaleSystemNamespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.ErrorMessage = "tailscale-system namespace not found"
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

// InstallOperator installs the Tailscale operator with OAuth credentials
func (c *Client) InstallOperator(ctx context.Context, opts *InstallOperatorOptions) error {
	if opts.Namespace == "" {
		opts.Namespace = TailscaleSystemNamespace
	}
	if opts.Image == "" {
		opts.Image = DefaultOperatorImage
	}

	// Validate OAuth credentials
	if opts.OAuthClientID == "" || opts.OAuthClientSecret == "" {
		return NewK8sError(ErrorTypeResourceInvalid,
			"OAuth client ID and secret are required for operator installation", nil)
	}

	// Check if operator is already installed
	status, err := c.GetOperatorStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check operator status: %w", err)
	}
	if status.Installed {
		return NewResourceConflictError("operator", OperatorDeploymentName, nil)
	}

	// Create namespace
	if err := c.createNamespace(ctx, opts.Namespace); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Create OAuth secret
	if err := c.createOAuthSecret(ctx, opts.Namespace, opts.OAuthClientID, opts.OAuthClientSecret); err != nil {
		return fmt.Errorf("failed to create OAuth secret: %w", err)
	}

	// Create service account
	if err := c.createServiceAccount(ctx, opts.Namespace); err != nil {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	// Create deployment
	if err := c.createOperatorDeployment(ctx, opts.Namespace, opts.Image); err != nil {
		return fmt.Errorf("failed to create operator deployment: %w", err)
	}

	// Wait for deployment to be ready
	if err := c.waitForOperatorReady(ctx, opts.Namespace, 5*time.Minute); err != nil {
		return NewOperatorInstallError("operator installation failed", err)
	}

	return nil
}

// UpgradeOperator upgrades the Tailscale operator to a new version
func (c *Client) UpgradeOperator(ctx context.Context, opts *UpgradeOperatorOptions) error {
	// Check if operator is installed
	status, err := c.GetOperatorStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check operator status: %w", err)
	}
	if !status.Installed {
		return NewOperatorNotFoundError(nil)
	}

	// Get current deployment
	deployment, err := c.clientset.AppsV1().Deployments(TailscaleSystemNamespace).Get(ctx, OperatorDeploymentName, metav1.GetOptions{})
	if err != nil {
		return NewResourceNotFoundError("deployment", OperatorDeploymentName, err)
	}

	// Update image
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return NewK8sError(ErrorTypeResourceInvalid, "deployment has no containers", nil)
	}

	currentImage := deployment.Spec.Template.Spec.Containers[0].Image
	if currentImage == opts.Image && !opts.Force {
		return NewK8sError(ErrorTypeResourceConflict,
			fmt.Sprintf("operator is already using image %s", opts.Image), nil)
	}

	// Update the image
	deployment.Spec.Template.Spec.Containers[0].Image = opts.Image

	// Update deployment
	_, err = c.clientset.AppsV1().Deployments(TailscaleSystemNamespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return NewK8sError(ErrorTypeOperatorUpgrade, "failed to update operator deployment", err)
	}

	// Wait for rollout to complete
	if err := c.waitForOperatorReady(ctx, TailscaleSystemNamespace, 5*time.Minute); err != nil {
		return NewK8sError(ErrorTypeOperatorUpgrade, "operator upgrade failed", err)
	}

	return nil
}

// createNamespace creates the tailscale-system namespace if it doesn't exist
func (c *Client) createNamespace(ctx context.Context, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace %s: %w", name, err)
	}

	return nil
}

// createOAuthSecret creates the OAuth secret for the operator
func (c *Client) createOAuthSecret(ctx context.Context, namespace, clientID, clientSecret string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "operator-oauth",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"client_id":     []byte(clientID),
			"client_secret": []byte(clientSecret),
		},
	}

	_, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create OAuth secret: %w", err)
	}

	return nil
}

// createServiceAccount creates the service account for the operator
func (c *Client) createServiceAccount(ctx context.Context, namespace string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OperatorServiceAccount,
			Namespace: namespace,
		},
	}

	_, err := c.clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	return nil
}

// createOperatorDeployment creates the operator deployment
func (c *Client) createOperatorDeployment(ctx context.Context, namespace, image string) error {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OperatorDeploymentName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "tailscale-operator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "tailscale-operator",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: OperatorServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  "operator",
							Image: image,
							Env: []corev1.EnvVar{
								{
									Name: "OPERATOR_OAUTH_CLIENT_ID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "operator-oauth",
											},
											Key: "client_id",
										},
									},
								},
								{
									Name: "OPERATOR_OAUTH_CLIENT_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "operator-oauth",
											},
											Key: "client_secret",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create operator deployment: %w", err)
	}

	return nil
}

// waitForOperatorReady waits for the operator deployment to be ready
func (c *Client) waitForOperatorReady(ctx context.Context, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, OperatorDeploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
			deployment.Status.ReadyReplicas > 0, nil
	})
}