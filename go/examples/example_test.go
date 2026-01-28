package examples_test

import (
	"context"
	"fmt"
	"log"

	"github.com/roma-glushko/testcontainers-envtest/go"
	"github.com/testcontainers/testcontainers-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Example() {
	ctx := context.Background()

	// Start an envtest container
	container, err := envtest.Run(ctx)
	if err != nil {
		log.Fatalf("failed to start envtest container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	// Get a REST config for the Kubernetes client
	cfg, err := container.GetRESTConfig(ctx)
	if err != nil {
		log.Fatalf("failed to get REST config: %v", err)
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create kubernetes client: %v", err)
	}

	// List namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list namespaces: %v", err)
	}

	fmt.Printf("Found %d namespaces\n", len(namespaces.Items))
	// Output: Found 4 namespaces
}

func Example_withVersion() {
	ctx := context.Background()

	// Start an envtest container with a specific Kubernetes version
	container, err := envtest.Run(ctx,
		envtest.WithKubernetesVersion("1.30.0"),
	)
	if err != nil {
		log.Fatalf("failed to start envtest container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	// Get the API server URL
	url, err := container.GetAPIServerURL(ctx)
	if err != nil {
		log.Fatalf("failed to get API server URL: %v", err)
	}

	fmt.Printf("Kubernetes version: %s\n", container.GetKubernetesVersion())
	fmt.Printf("API Server URL: %s\n", url)
}
