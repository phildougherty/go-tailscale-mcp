package k8s

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ProxyClass represents a Tailscale ProxyClass resource
type ProxyClass struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       ProxyClassSpec    `json:"spec"`
	Status     *ProxyClassStatus `json:"status,omitempty"`
}

type ProxyClassSpec struct {
	ProxyImage     string            `json:"proxyImage,omitempty"`
	StatefulSet    *StatefulSetSpec  `json:"statefulSet,omitempty"`
	TailscaleConfig map[string]string `json:"tailscaleConfig,omitempty"`
}

type StatefulSetSpec struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Pod         *PodSpec          `json:"pod,omitempty"`
}

type PodSpec struct {
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	TailscaleContainer *TailscaleContainer `json:"tailscaleContainer,omitempty"`
}

type TailscaleContainer struct {
	Env       []corev1.EnvVar           `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ProxyClassStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ProxyGroup represents a Tailscale ProxyGroup resource
type ProxyGroup struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       ProxyGroupSpec   `json:"spec"`
	Status     *ProxyGroupStatus `json:"status,omitempty"`
}

type ProxyGroupSpec struct {
	Type        string   `json:"type"` // "egress" or "ingress"
	Replicas    *int32   `json:"replicas,omitempty"`
	ProxyClass  string   `json:"proxyClass,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type ProxyGroupStatus struct {
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
	Replicas       int32              `json:"replicas"`
	ReadyReplicas  int32              `json:"readyReplicas"`
}

// Connector represents a Tailscale Connector resource
type Connector struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       ConnectorSpec    `json:"spec"`
	Status     *ConnectorStatus `json:"status,omitempty"`
}

type ConnectorSpec struct {
	Hostname       string            `json:"hostname,omitempty"`
	ProxyClass     string            `json:"proxyClass,omitempty"`
	SubnetRouter   *SubnetRouterSpec `json:"subnetRouter,omitempty"`
	ExitNode       bool              `json:"exitNode,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
}

type SubnetRouterSpec struct {
	AdvertiseRoutes []string `json:"advertiseRoutes"`
}

type ConnectorStatus struct {
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	Hostname        string             `json:"hostname,omitempty"`
	TailscaleIPs    []string           `json:"tailscaleIPs,omitempty"`
}

// DNSConfig represents a Tailscale DNSConfig resource
type DNSConfig struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       DNSConfigSpec    `json:"spec"`
	Status     *DNSConfigStatus `json:"status,omitempty"`
}

type DNSConfigSpec struct {
	MagicDNS    bool              `json:"magicDNS"`
	Nameservers []NameserverSpec  `json:"nameservers,omitempty"`
}

type NameserverSpec struct {
	IP string `json:"ip"`
}

type DNSConfigStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Resource GVRs (GroupVersionResource)
var (
	ProxyClassGVR = schema.GroupVersionResource{
		Group:    "tailscale.com",
		Version:  "v1alpha1",
		Resource: "proxyclasses",
	}
	ProxyGroupGVR = schema.GroupVersionResource{
		Group:    "tailscale.com",
		Version:  "v1alpha1",
		Resource: "proxygroups",
	}
	ConnectorGVR = schema.GroupVersionResource{
		Group:    "tailscale.com",
		Version:  "v1alpha1",
		Resource: "connectors",
	}
	DNSConfigGVR = schema.GroupVersionResource{
		Group:    "tailscale.com",
		Version:  "v1alpha1",
		Resource: "dnsconfigs",
	}
)

// ResourceManager handles Tailscale custom resources
type ResourceManager struct {
	client       *Client
	dynamicClient dynamic.Interface
}

// NewResourceManager creates a new resource manager
func NewResourceManager(client *Client) (*ResourceManager, error) {
	dynamicClient, err := dynamic.NewForConfig(client.config)
	if err != nil {
		return nil, NewConnectivityError("failed to create dynamic client", err)
	}

	return &ResourceManager{
		client:        client,
		dynamicClient: dynamicClient,
	}, nil
}

// CreateProxyClass creates a ProxyClass resource
func (rm *ResourceManager) CreateProxyClass(ctx context.Context, proxyClass *ProxyClass) error {
	proxyClass.APIVersion = "tailscale.com/v1alpha1"
	proxyClass.Kind = "ProxyClass"

	unstructuredObj, err := toUnstructured(proxyClass)
	if err != nil {
		return NewK8sError(ErrorTypeResourceInvalid, "failed to convert ProxyClass to unstructured", err)
	}

	_, err = rm.dynamicClient.Resource(ProxyClassGVR).Namespace(proxyClass.Metadata.Namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("ProxyClass", proxyClass.Metadata.Name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create ProxyClass", err)
	}

	return nil
}

// ListProxyClasses lists all ProxyClass resources in a namespace
func (rm *ResourceManager) ListProxyClasses(ctx context.Context, namespace string) ([]ProxyClass, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	unstructuredList, err := rm.dynamicClient.Resource(ProxyClassGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, NewConnectivityError("failed to list ProxyClasses", err)
	}

	var proxyClasses []ProxyClass
	for _, item := range unstructuredList.Items {
		var pc ProxyClass
		if err := fromUnstructured(&item, &pc); err != nil {
			continue // Skip invalid items
		}
		proxyClasses = append(proxyClasses, pc)
	}

	return proxyClasses, nil
}

// DeleteProxyClass deletes a ProxyClass resource
func (rm *ResourceManager) DeleteProxyClass(ctx context.Context, namespace, name string) error {
	err := rm.dynamicClient.Resource(ProxyClassGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return NewResourceNotFoundError("ProxyClass", name, err)
		}
		return NewK8sError(ErrorTypeUnknown, "failed to delete ProxyClass", err)
	}

	return nil
}

// CreateProxyGroup creates a ProxyGroup resource
func (rm *ResourceManager) CreateProxyGroup(ctx context.Context, proxyGroup *ProxyGroup) error {
	proxyGroup.APIVersion = "tailscale.com/v1alpha1"
	proxyGroup.Kind = "ProxyGroup"

	unstructuredObj, err := toUnstructured(proxyGroup)
	if err != nil {
		return NewK8sError(ErrorTypeResourceInvalid, "failed to convert ProxyGroup to unstructured", err)
	}

	_, err = rm.dynamicClient.Resource(ProxyGroupGVR).Namespace(proxyGroup.Metadata.Namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("ProxyGroup", proxyGroup.Metadata.Name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create ProxyGroup", err)
	}

	return nil
}

// GetProxyGroupStatus gets the status of a ProxyGroup resource
func (rm *ResourceManager) GetProxyGroupStatus(ctx context.Context, namespace, name string) (*ProxyGroupStatus, error) {
	unstructuredObj, err := rm.dynamicClient.Resource(ProxyGroupGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, NewResourceNotFoundError("ProxyGroup", name, err)
		}
		return nil, NewConnectivityError("failed to get ProxyGroup", err)
	}

	var proxyGroup ProxyGroup
	if err := fromUnstructured(unstructuredObj, &proxyGroup); err != nil {
		return nil, NewK8sError(ErrorTypeResourceInvalid, "failed to parse ProxyGroup", err)
	}

	return proxyGroup.Status, nil
}

// ScaleProxyGroup scales a ProxyGroup resource
func (rm *ResourceManager) ScaleProxyGroup(ctx context.Context, namespace, name string, replicas int32) error {
	unstructuredObj, err := rm.dynamicClient.Resource(ProxyGroupGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return NewResourceNotFoundError("ProxyGroup", name, err)
		}
		return NewConnectivityError("failed to get ProxyGroup", err)
	}

	// Update replicas in spec
	if err := unstructured.SetNestedField(unstructuredObj.Object, int64(replicas), "spec", "replicas"); err != nil {
		return NewK8sError(ErrorTypeResourceInvalid, "failed to set replicas", err)
	}

	_, err = rm.dynamicClient.Resource(ProxyGroupGVR).Namespace(namespace).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return NewK8sError(ErrorTypeUnknown, "failed to scale ProxyGroup", err)
	}

	return nil
}

// CreateConnector creates a Connector resource
func (rm *ResourceManager) CreateConnector(ctx context.Context, connector *Connector) error {
	connector.APIVersion = "tailscale.com/v1alpha1"
	connector.Kind = "Connector"

	unstructuredObj, err := toUnstructured(connector)
	if err != nil {
		return NewK8sError(ErrorTypeResourceInvalid, "failed to convert Connector to unstructured", err)
	}

	_, err = rm.dynamicClient.Resource(ConnectorGVR).Namespace(connector.Metadata.Namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("Connector", connector.Metadata.Name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create Connector", err)
	}

	return nil
}

// CreateDNSConfig creates a DNSConfig resource
func (rm *ResourceManager) CreateDNSConfig(ctx context.Context, dnsConfig *DNSConfig) error {
	dnsConfig.APIVersion = "tailscale.com/v1alpha1"
	dnsConfig.Kind = "DNSConfig"

	unstructuredObj, err := toUnstructured(dnsConfig)
	if err != nil {
		return NewK8sError(ErrorTypeResourceInvalid, "failed to convert DNSConfig to unstructured", err)
	}

	_, err = rm.dynamicClient.Resource(DNSConfigGVR).Namespace(dnsConfig.Metadata.Namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("DNSConfig", dnsConfig.Metadata.Name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create DNSConfig", err)
	}

	return nil
}

// CreateTailscaleIngress creates a Tailscale ingress using a standard Kubernetes Ingress with Tailscale annotations
func (rm *ResourceManager) CreateTailscaleIngress(ctx context.Context, namespace, name, hostname, serviceName string, servicePort int32) error {
	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"tailscale.com/expose":   "true",
				"tailscale.com/hostname": hostname,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: servicePort,
											},
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

	_, err := rm.client.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("Ingress", name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create Tailscale ingress", err)
	}

	return nil
}

// CreateEgressService creates an egress service for Tailscale
func (rm *ResourceManager) CreateEgressService(ctx context.Context, namespace, name, externalHostname string, port int32) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"tailscale.com/expose": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: externalHostname,
			Ports: []corev1.ServicePort{
				{
					Port:     port,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := rm.client.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return NewResourceConflictError("Service", name, err)
		}
		return NewK8sError(ErrorTypeResourceInvalid, "failed to create egress service", err)
	}

	return nil
}

// Helper functions for converting between structured and unstructured objects
func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var unstructuredObj unstructured.Unstructured
	if err := json.Unmarshal(data, &unstructuredObj); err != nil {
		return nil, err
	}

	return &unstructuredObj, nil
}

func fromUnstructured(obj *unstructured.Unstructured, target interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}