package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	envtest "github.com/roma-glushko/testcontainers-envtest/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestConfigMapHashController(t *testing.T) {
	ctx := t.Context()

	k8s, err := envtest.Run(ctx)
	require.NoError(t, err)

	defer func() {
		err := testcontainers.TerminateContainer(k8s)
		require.NoError(t, err)
	}()

	// Get REST config from container
	restConfig, err := k8s.RESTConfig(ctx)
	require.NoError(t, err)

	// Create manager with the test cluster config
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{})
	require.NoError(t, err)

	// Setup controller
	err = (&ConfigMapHashReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)

	require.NoError(t, err)

	// Start manager in background
	mgrCtx, mgrCancel := context.WithCancel(ctx)
	defer mgrCancel()

	go func() {
		if err := mgr.Start(mgrCtx); err != nil {
			t.Logf("Manager stopped: %v", err)
		}
	}()

	// Wait for cache to sync
	require.Eventually(t, func() bool {
		return mgr.GetCache().WaitForCacheSync(ctx)
	}, 30*time.Second, 100*time.Millisecond, "Cache failed to sync")

	k8sClient := mgr.GetClient()

	// Create a ConfigMap with the watch label
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "default",
			Labels: map[string]string{
				WatchLabel: "true",
			},
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err = k8sClient.Create(ctx, configMap)
	require.NoError(t, err)

	// Wait for the controller to add hash annotations
	var updatedCM corev1.ConfigMap
	require.Eventually(t, func() bool {
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      "test-configmap",
			Namespace: "default",
		}, &updatedCM)

		if err != nil {
			return false
		}

		// Check if hash annotations exist
		hashKeys := GetHashAnnotationKeys(updatedCM.Annotations)
		return len(hashKeys) == 2
	}, 5*time.Second, 50*time.Millisecond, "ConfigMap was not updated with hash annotations")

	// Verify hash values
	expectedHash1 := sha256Hash("value1")
	expectedHash2 := sha256Hash("value2")

	assert.Equal(t, expectedHash1, updatedCM.Annotations[AnnotationPrefix+"key1"])
	assert.Equal(t, expectedHash2, updatedCM.Annotations[AnnotationPrefix+"key2"])

	// Test: ConfigMap without label should not be processed
	configMapNoLabel := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-no-label",
			Namespace: "default",
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	err = k8sClient.Create(ctx, configMapNoLabel)
	require.NoError(t, err, "Failed to create ConfigMap without label")

	// Wait a bit and verify no annotations were added
	time.Sleep(500 * time.Millisecond)

	var cmNoLabel corev1.ConfigMap
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "test-configmap-no-label",
		Namespace: "default",
	}, &cmNoLabel)
	require.NoError(t, err)
	require.Empty(t, GetHashAnnotationKeys(cmNoLabel.Annotations), "ConfigMap without label should not have hash annotations")
}

func sha256Hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
