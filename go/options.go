package envtest

// config holds the configuration for the envtest container
type config struct {
	image             string
	kubernetesVersion string
}

// Option is a functional option for configuring the envtest container
type Option func(*config)

// WithImage sets a custom Docker image for the envtest container.
// By default, it uses ghcr.io/roma-glushko/testcontainers-envtest:latest
func WithImage(image string) Option {
	return func(c *config) {
		c.image = image
	}
}

// WithKubernetesVersion sets the Kubernetes version to use.
// This will automatically select the appropriate image tag.
// Supported versions: 1.27, 1.28, 1.29, 1.30, 1.31
func WithKubernetesVersion(version string) Option {
	return func(c *config) {
		c.kubernetesVersion = version
	}
}
