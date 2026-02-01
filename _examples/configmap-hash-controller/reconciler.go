package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"maps"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"
)

const (
	// WatchLabel is the label that ConfigMaps must have to be watched by this controller
	WatchLabel = "configmap-hash-controller.example.io/enabled"
	// AnnotationPrefix is the prefix for hash annotations
	AnnotationPrefix = "configmap-hash-controller.example.io/hash-"
)

// ConfigMapHashReconciler reconciles ConfigMaps and adds hash annotations
type ConfigMapHashReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles ConfigMap reconciliation
func (r *ConfigMapHashReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	var configMap corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &configMap); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling ConfigMap", "name", configMap.Name, "namespace", configMap.Namespace)

	// Calculate hashes for each key
	newAnnotations := make(map[string]string)

	// Copy existing non-hash annotations
	for k, v := range configMap.Annotations {
		if len(k) < len(AnnotationPrefix) || k[:len(AnnotationPrefix)] != AnnotationPrefix {
			newAnnotations[k] = v
		}
	}

	// Calculate and add hash annotations for each data key
	for key, value := range configMap.Data {
		hash := calculateHash(value)
		annotationKey := AnnotationPrefix + key
		newAnnotations[annotationKey] = hash
	}

	if !maps.Equal(newAnnotations, configMap.Annotations) {
		configMap.Annotations = newAnnotations
		if err := r.Update(ctx, &configMap); err != nil {
			log.Error(err, "Failed to update ConfigMap annotations")
			return ctrl.Result{}, err
		}

		log.Info("Updated ConfigMap annotations with hashes", "name", configMap.Name)
	}

	return ctrl.Result{}, nil
}

// calculateHash computes SHA256 hash of the given value
func calculateHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// SetupWithManager sets up the controller with the Manager
func (r *ConfigMapHashReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a predicate that filters ConfigMaps with the watch label
	labelPredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		labels := obj.GetLabels()
		if labels == nil {
			return false
		}
		value, exists := labels[WatchLabel]
		return exists && value == "true"
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(labelPredicate).
		Complete(r)
}

// GetHashAnnotationKeys returns sorted list of hash annotation keys for testing
func GetHashAnnotationKeys(annotations map[string]string) []string {
	var keys []string

	for k := range annotations {
		if len(k) >= len(AnnotationPrefix) && k[:len(AnnotationPrefix)] == AnnotationPrefix {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	return keys
}
