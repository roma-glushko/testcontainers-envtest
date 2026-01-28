package envtest

import (
	"testing"
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
			if got != tt.want {
				t.Errorf("replaceServerURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   int
	}{
		{
			name:   "found at beginning",
			s:      "hello world",
			substr: "hello",
			want:   0,
		},
		{
			name:   "found in middle",
			s:      "hello world",
			substr: "wor",
			want:   6,
		},
		{
			name:   "found at end",
			s:      "hello world",
			substr: "world",
			want:   6,
		},
		{
			name:   "not found",
			s:      "hello world",
			substr: "foo",
			want:   -1,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "foo",
			want:   -1,
		},
		{
			name:   "empty substr",
			s:      "hello",
			substr: "",
			want:   0,
		},
		{
			name:   "substr longer than string",
			s:      "hi",
			substr: "hello",
			want:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findSubstring(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("findSubstring(%q, %q) = %d, want %d", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestWithImage(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	customImage := "my-registry/envtest:custom"
	WithImage(customImage)(cfg)

	if cfg.image != customImage {
		t.Errorf("WithImage() set image to %q, want %q", cfg.image, customImage)
	}
}

func TestWithKubernetesVersion(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	version := "1.30.0"
	WithKubernetesVersion(version)(cfg)

	if cfg.kubernetesVersion != version {
		t.Errorf("WithKubernetesVersion() set version to %q, want %q", cfg.kubernetesVersion, version)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	if cfg.image != DefaultImage {
		t.Errorf("default image = %q, want %q", cfg.image, DefaultImage)
	}

	if cfg.kubernetesVersion != DefaultKubernetesVersion {
		t.Errorf("default kubernetes version = %q, want %q", cfg.kubernetesVersion, DefaultKubernetesVersion)
	}
}

func TestConstants(t *testing.T) {
	if DefaultImage == "" {
		t.Error("DefaultImage should not be empty")
	}

	if DefaultKubernetesVersion == "" {
		t.Error("DefaultKubernetesVersion should not be empty")
	}

	if DefaultAPIServerPort == "" {
		t.Error("DefaultAPIServerPort should not be empty")
	}

	if KubeconfigPath == "" {
		t.Error("KubeconfigPath should not be empty")
	}

	// Verify port is a valid number
	if DefaultAPIServerPort != "6443" {
		t.Errorf("DefaultAPIServerPort = %q, expected 6443", DefaultAPIServerPort)
	}
}
