package k8s

import (
	"context"
	"fmt"

	"github.com/homecraft/backend/pkg/apis/homecraft/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client
type Client struct {
	config     *rest.Config
	clientset  *kubernetes.Clientset
	restClient *rest.RESTClient
	scheme     *runtime.Scheme
}

// NewClient creates a new Kubernetes client
// It tries to use in-cluster config first, then falls back to kubeconfig
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Create scheme and add our types
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add types to scheme: %w", err)
	}

	// Create REST client for custom resources
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   v1alpha1.GroupName,
		Version: v1alpha1.Version,
	}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}

	return &Client{
		config:     config,
		clientset:  clientset,
		restClient: restClient,
		scheme:     scheme,
	}, nil
}

// CreateMinecraftServer creates a new MinecraftServer custom resource
func (c *Client) CreateMinecraftServer(ctx context.Context, namespace string, server *v1alpha1.MinecraftServer) (*v1alpha1.MinecraftServer, error) {
	result := &v1alpha1.MinecraftServer{}
	err := c.restClient.Post().
		Namespace(namespace).
		Resource("minecraftservers").
		Body(server).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, fmt.Errorf("failed to create MinecraftServer: %w", err)
	}
	return result, nil
}

// GetMinecraftServer retrieves a MinecraftServer by name
func (c *Client) GetMinecraftServer(ctx context.Context, namespace, name string) (*v1alpha1.MinecraftServer, error) {
	result := &v1alpha1.MinecraftServer{}
	err := c.restClient.Get().
		Namespace(namespace).
		Resource("minecraftservers").
		Name(name).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, fmt.Errorf("failed to get MinecraftServer: %w", err)
	}
	return result, nil
}

// ListMinecraftServers lists all MinecraftServers in a namespace
func (c *Client) ListMinecraftServers(ctx context.Context, namespace string) (*v1alpha1.MinecraftServerList, error) {
	result := &v1alpha1.MinecraftServerList{}
	err := c.restClient.Get().
		Namespace(namespace).
		Resource("minecraftservers").
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, fmt.Errorf("failed to list MinecraftServers: %w", err)
	}
	return result, nil
}

// DeleteMinecraftServer deletes a MinecraftServer by name
func (c *Client) DeleteMinecraftServer(ctx context.Context, namespace, name string) error {
	err := c.restClient.Delete().
		Namespace(namespace).
		Resource("minecraftservers").
		Name(name).
		Body(&metav1.DeleteOptions{}).
		Do(ctx).
		Error()
	if err != nil {
		return fmt.Errorf("failed to delete MinecraftServer: %w", err)
	}
	return nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetClusterMemoryResources fetches cluster memory capacity and usage
func (c *Client) GetClusterMemoryResources(ctx context.Context) (totalMemory, allocatedMemory, availableMemory int64, err error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	var total int64
	var allocated int64

	for _, node := range nodes.Items {
		// Get allocatable memory (what's available for pods)
		if memory, ok := node.Status.Allocatable["memory"]; ok {
			total += memory.Value()
		}
	}

	// Get all pods to calculate allocated memory
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		// Skip completed/failed pods
		if pod.Status.Phase == "Succeeded" || pod.Status.Phase == "Failed" {
			continue
		}

		for _, container := range pod.Spec.Containers {
			if memory, ok := container.Resources.Requests["memory"]; ok {
				allocated += memory.Value()
			}
		}
	}

	available := total - allocated
	return total, allocated, available, nil
}

// CheckMemoryAvailability checks if requested memory is available in the cluster
func (c *Client) CheckMemoryAvailability(ctx context.Context, requestedMemory int64) (bool, string, error) {
	_, _, available, err := c.GetClusterMemoryResources(ctx)
	if err != nil {
		return false, "", err
	}

	if requestedMemory > available {
		return false, fmt.Sprintf("insufficient memory: requested %s, available %s",
			bytesToHumanReadable(requestedMemory),
			bytesToHumanReadable(available)), nil
	}

	return true, "", nil
}

// bytesToHumanReadable converts bytes to human-readable format
func bytesToHumanReadable(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
