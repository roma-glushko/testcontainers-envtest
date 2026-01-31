package envtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceServerURL(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		newURL     string
		want       string
	}{
		{
			name: "replace localhost URL",
			kubeconfig: `apiVersion: v1
clusters:
- cluster:
    server: https://localhost:6443
  name: envtest
`,
			newURL: "https://192.168.1.100:32768",
			want: `apiVersion: v1
clusters:
- cluster:
    server: https://192.168.1.100:32768
  name: envtest
`,
		},
		{
			name: "replace 127.0.0.1 URL",
			kubeconfig: `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: envtest
`,
			newURL: "https://host.docker.internal:45678",
			want: `apiVersion: v1
clusters:
- cluster:
    server: https://host.docker.internal:45678
  name: envtest
`,
		},
		{
			name: "no match - different host",
			kubeconfig: `apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.default:443
  name: envtest
`,
			newURL: "https://192.168.1.100:32768",
			want: `apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.default:443
  name: envtest
`,
		},
		{
			name:       "empty kubeconfig",
			kubeconfig: "",
			newURL:     "https://localhost:1234",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceServerURL(tt.kubeconfig, tt.newURL)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestConstants(t *testing.T) {
	require.NotEmpty(t, DefaultImage)
	require.NotEmpty(t, DefaultKubernetesVersion)
	require.NotEmpty(t, DefaultAPIServerPort)
	require.NotEmpty(t, KubeconfigPath)
	require.Equal(t, "6443", DefaultAPIServerPort)
}
