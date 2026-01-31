package envtest

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithImage(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	customImage := "my-registry/envtest:custom"
	WithImage(customImage)(cfg)

	require.Equal(t, customImage, cfg.image)
}

func TestWithKubernetesVersion(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	version := "1.30.0"
	WithKubernetesVersion(version)(cfg)

	require.Equal(t, version, cfg.kubernetesVersion)
}

func TestDefaultConfig(t *testing.T) {
	cfg := &config{
		image:             DefaultImage,
		kubernetesVersion: DefaultKubernetesVersion,
	}

	require.Equal(t, DefaultImage, cfg.image)
	require.Equal(t, DefaultKubernetesVersion, cfg.kubernetesVersion)
}
