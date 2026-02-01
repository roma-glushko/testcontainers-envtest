package envtest_test

import (
	"context"
	"testing"
	"time"

	envtest "github.com/roma-glushko/testcontainers-envtest/go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestEnvtestContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	c, err := envtest.Run(ctx)
	require.NoError(t, err, "failed to start envtest container")

	defer func() {
		err := testcontainers.TerminateContainer(c)
		require.NoError(t, err, "failed to terminate container")
	}()

	t.Run("APIServerURL", func(t *testing.T) {
		url, err := c.APIServerURL(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, url)

		t.Logf("API Server URL: %s", url)
	})

	t.Run("Kubeconfig", func(t *testing.T) {
		kubeconfig, err := c.Kubeconfig(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, kubeconfig)
		require.Greater(t, len(kubeconfig), 100, "kubeconfig seems too short")

		t.Logf("Kubeconfig length: %d bytes", len(kubeconfig))
	})

	t.Run("RESTConfig", func(t *testing.T) {
		cfg, err := c.RESTConfig(ctx)
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
		cfg, err := c.RESTConfig(ctx)
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

	container, err := envtest.Run(ctx, envtest.WithKubernetesVersion("1.35.0"))
	require.NoError(t, err)

	defer func() {
		err := testcontainers.TerminateContainer(container)
		require.NoError(t, err)
	}()

	require.Equal(t, "1.35.0", container.KubernetesVersion())
}

func BenchmarkContainerLifecycle(b *testing.B) {
	b.Run("envtest", func(b *testing.B) {
		ctx := b.Context()

		for b.Loop() {
			container, err := envtest.Run(ctx)
			require.NoError(b, err)

			_ = testcontainers.TerminateContainer(container)
		}

		b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(b.N), "ms/op")
	})

	b.Run("k3s", func(b *testing.B) {
		ctx := b.Context()

		for b.Loop() {
			container, err := k3s.Run(ctx, "rancher/k3s:v1.31.2-k3s1")
			require.NoError(b, err)

			_ = testcontainers.TerminateContainer(container)
		}

		b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(b.N), "ms/op")
	})
}
