package envtest_test

import (
	"context"
	"testing"
	"time"

	"github.com/roma-glushko/testcontainers-envtest/go"
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestEnvtestContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, err := envtest.Run(ctx)
	if err != nil {
		t.Fatalf("failed to start envtest container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// Test GetAPIServerURL
	t.Run("GetAPIServerURL", func(t *testing.T) {
		url, err := container.GetAPIServerURL(ctx)
		if err != nil {
			t.Fatalf("failed to get API server URL: %v", err)
		}
		if url == "" {
			t.Fatal("API server URL is empty")
		}
		t.Logf("API Server URL: %s", url)
	})

	// Test GetKubeconfig
	t.Run("GetKubeconfig", func(t *testing.T) {
		kubeconfig, err := container.GetKubeconfig(ctx)
		if err != nil {
			t.Fatalf("failed to get kubeconfig: %v", err)
		}
		if kubeconfig == "" {
			t.Fatal("kubeconfig is empty")
		}
		if len(kubeconfig) < 100 {
			t.Fatalf("kubeconfig seems too short: %d bytes", len(kubeconfig))
		}
		t.Logf("Kubeconfig length: %d bytes", len(kubeconfig))
	})

	// Test GetRESTConfig and using it with a real client
	t.Run("GetRESTConfig", func(t *testing.T) {
		cfg, err := container.GetRESTConfig(ctx)
		if err != nil {
			t.Fatalf("failed to get REST config: %v", err)
		}
		if cfg == nil {
			t.Fatal("REST config is nil")
		}

		// Create a kubernetes client
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("failed to create kubernetes client: %v", err)
		}

		// List namespaces to verify the connection works
		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("failed to list namespaces: %v", err)
		}

		t.Logf("Found %d namespaces", len(namespaces.Items))
		for _, ns := range namespaces.Items {
			t.Logf("  - %s", ns.Name)
		}
	})

	// Test creating a resource
	t.Run("CreateNamespace", func(t *testing.T) {
		cfg, err := container.GetRESTConfig(ctx)
		if err != nil {
			t.Fatalf("failed to get REST config: %v", err)
		}

		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("failed to create kubernetes client: %v", err)
		}

		// Create a test namespace
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		}

		created, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create namespace: %v", err)
		}

		t.Logf("Created namespace: %s", created.Name)

		// Verify the namespace exists
		got, err := clientset.CoreV1().Namespaces().Get(ctx, "test-namespace", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("failed to get namespace: %v", err)
		}

		if got.Name != "test-namespace" {
			t.Fatalf("expected namespace name 'test-namespace', got '%s'", got.Name)
		}
	})
}

func TestEnvtestContainerWithKubernetesVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, err := envtest.Run(ctx, envtest.WithKubernetesVersion("1.30.0"))
	if err != nil {
		t.Fatalf("failed to start envtest container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	version := container.GetKubernetesVersion()
	if version != "1.30.0" {
		t.Fatalf("expected kubernetes version '1.30.0', got '%s'", version)
	}
}
