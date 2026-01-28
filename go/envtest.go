// Package envtest provides a testcontainers module for running envtest
// (Kubernetes API server + etcd) in a container for testing Kubernetes
// controllers and operators.
package envtest

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// DefaultImage is the default Docker image for envtest
	DefaultImage = "ghcr.io/roma-glushko/testcontainers-envtest:latest"

	// DefaultKubernetesVersion is the default Kubernetes version
	DefaultKubernetesVersion = "1.31.0"

	// DefaultAPIServerPort is the default port for the Kubernetes API server
	DefaultAPIServerPort = "6443"

	// KubeconfigPath is the path to the kubeconfig inside the container
	KubeconfigPath = "/tmp/kubeconfig"
)

// EnvtestContainer represents an envtest container instance
type EnvtestContainer struct {
	testcontainers.Container
	kubernetesVersion string
}

// Run creates and starts an envtest container with the given options
func Run(ctx context.Context, opts ...Option) (*EnvtestContainer, error) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// If a specific kubernetes version is requested, use the versioned image tag
	image := cfg.image
	if cfg.kubernetesVersion != DefaultKubernetesVersion && cfg.image == DefaultImage {
		image = fmt.Sprintf("ghcr.io/roma-glushko/testcontainers-envtest:v%s", cfg.kubernetesVersion)
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{DefaultAPIServerPort + "/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(DefaultAPIServerPort+"/tcp"),
			wait.ForLog("Envtest is ready!"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start envtest container: %w", err)
	}

	return &EnvtestContainer{
		Container:         container,
		kubernetesVersion: cfg.kubernetesVersion,
	}, nil
}

// GetKubeconfig returns the kubeconfig YAML content for connecting to the API server
func (c *EnvtestContainer) GetKubeconfig(ctx context.Context) (string, error) {
	// Read the kubeconfig from the container
	reader, err := c.CopyFileFromContainer(ctx, KubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy kubeconfig from container: %w", err)
	}
	defer reader.Close()

	// Read all content
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := reader.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}

	// The kubeconfig has localhost as the server, we need to replace it
	// with the actual container host and mapped port
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := c.MappedPort(ctx, DefaultAPIServerPort+"/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Parse and modify the kubeconfig
	kubeconfig := string(buf)
	// Replace the server URL
	kubeconfig = replaceServerURL(kubeconfig, fmt.Sprintf("https://%s:%s", host, port.Port()))

	return kubeconfig, nil
}

// GetAPIServerURL returns the URL of the Kubernetes API server
func (c *EnvtestContainer) GetAPIServerURL(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := c.MappedPort(ctx, DefaultAPIServerPort+"/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	return fmt.Sprintf("https://%s:%s", host, port.Port()), nil
}

// GetRESTConfig returns a *rest.Config configured for the envtest API server.
// This config can be used with client-go or controller-runtime clients.
func (c *EnvtestContainer) GetRESTConfig(ctx context.Context) (*rest.Config, error) {
	kubeconfig, err := c.GetKubeconfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	return config, nil
}

// GetKubernetesVersion returns the Kubernetes version of the envtest container
func (c *EnvtestContainer) GetKubernetesVersion() string {
	return c.kubernetesVersion
}

// replaceServerURL replaces the server URL in a kubeconfig string
func replaceServerURL(kubeconfig, newURL string) string {
	// Simple string replacement for the server URL
	// The kubeconfig format has "server: https://localhost:PORT"
	result := kubeconfig
	for _, oldHost := range []string{"localhost", "127.0.0.1"} {
		oldURL := fmt.Sprintf("server: https://%s:", oldHost)
		if idx := findSubstring(result, oldURL); idx >= 0 {
			// Find the end of the line
			endIdx := idx + len(oldURL)
			for endIdx < len(result) && result[endIdx] != '\n' && result[endIdx] != '\r' {
				endIdx++
			}
			result = result[:idx] + "server: " + newURL + result[endIdx:]
			break
		}
	}
	return result
}

// findSubstring returns the index of substr in s, or -1 if not found
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
