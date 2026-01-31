package envtest_test

import (
	"context"
	"testing"
	"time"

	envtest "github.com/roma-glushko/testcontainers-envtest/go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestEnvtestContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	c, err := envtest.Run(ctx)
	require.NoError(t, err, "failed to start envtest container")

	defer func() {
		err := testcontainers.TerminateContainer(c)
		require.NoError(t, err, "failed to terminate container")
	}()

	t.Run("GetAPIServerURL", func(t *testing.T) {
		url, err := c.GetAPIServerURL(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, url)

		t.Logf("API Server URL: %s", url)
	})

	t.Run("GetKubeconfig", func(t *testing.T) {
		kubeconfig, err := c.GetKubeconfig(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, kubeconfig)
		require.Greater(t, len(kubeconfig), 100, "kubeconfig seems too short")

		t.Logf("Kubeconfig length: %d bytes", len(kubeconfig))
	})

	t.Run("GetRESTConfig", func(t *testing.T) {
		cfg, err := c.GetRESTConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		clientset, err := kubernetes.NewForConfig(cfg)
		require.NoError(t, err)

		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, namespaces.Items)

		t.Logf("Found %d namespaces", len(namespaces.Items))
		for _, ns := range namespaces.Items {
			t.Logf("  - %s", ns.Name)
		}
	})

	t.Run("CreateNamespace", func(t *testing.T) {
		cfg, err := c.GetRESTConfig(ctx)
		require.NoError(t, err)

		clientset, err := kubernetes.NewForConfig(cfg)
		require.NoError(t, err)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		}

		created, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		require.NoError(t, err)

		t.Logf("Created namespace: %s", created.Name)

		got, err := clientset.CoreV1().Namespaces().Get(ctx, "test-namespace", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, "test-namespace", got.Name)
	})
}

func TestEnvtestContainerWithKubernetesVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	container, err := envtest.Run(ctx, envtest.WithKubernetesVersion("1.30.0"))
	require.NoError(t, err, "failed to start envtest container")

	defer func() {
		err := testcontainers.TerminateContainer(container)
		require.NoError(t, err)
	}()

	require.Equal(t, "1.30.0", container.GetKubernetesVersion())
}
