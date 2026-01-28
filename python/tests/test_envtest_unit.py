"""Unit tests for EnvtestContainer that don't require Docker."""

import re

import pytest

from testcontainers_envtest import EnvtestContainer


class TestEnvtestContainerUnit:
    """Unit tests for EnvtestContainer configuration."""

    def test_default_image(self) -> None:
        """Test that the default image is set correctly."""
        assert EnvtestContainer.DEFAULT_IMAGE == "ghcr.io/roma-glushko/testcontainers-envtest:latest"

    def test_default_kubernetes_version(self) -> None:
        """Test that the default Kubernetes version is set correctly."""
        assert EnvtestContainer.DEFAULT_KUBERNETES_VERSION == "1.31.0"

    def test_api_server_port(self) -> None:
        """Test that the API server port is set correctly."""
        assert EnvtestContainer.API_SERVER_PORT == 6443

    def test_kubeconfig_path(self) -> None:
        """Test that the kubeconfig path is set correctly."""
        assert EnvtestContainer.KUBECONFIG_PATH == "/tmp/kubeconfig"

    def test_default_initialization(self) -> None:
        """Test container initialization with defaults."""
        container = EnvtestContainer()

        assert container.kubernetes_version == EnvtestContainer.DEFAULT_KUBERNETES_VERSION
        assert container.image == EnvtestContainer.DEFAULT_IMAGE

    def test_custom_kubernetes_version(self) -> None:
        """Test container initialization with custom Kubernetes version."""
        container = EnvtestContainer(kubernetes_version="1.30.0")

        assert container.kubernetes_version == "1.30.0"
        # Should use versioned image tag
        assert "v1.30.0" in container.image

    def test_custom_image(self) -> None:
        """Test container initialization with custom image."""
        custom_image = "my-registry/envtest:custom"
        container = EnvtestContainer(image=custom_image)

        assert container.image == custom_image

    def test_custom_image_overrides_version(self) -> None:
        """Test that custom image takes precedence over version-based image."""
        custom_image = "my-registry/envtest:custom"
        container = EnvtestContainer(image=custom_image, kubernetes_version="1.30.0")

        # Custom image should be used, not version-based image
        assert container.image == custom_image
        # But kubernetes_version property should still reflect the set version
        assert container.kubernetes_version == "1.30.0"

    def test_exposed_ports(self) -> None:
        """Test that the API server port is exposed."""
        container = EnvtestContainer()

        # Check that exposed_ports contains the API server port
        exposed_ports = container.exposed_ports
        assert EnvtestContainer.API_SERVER_PORT in exposed_ports


class TestKubeconfigParsing:
    """Tests for kubeconfig URL replacement logic."""

    def test_replace_localhost_url(self) -> None:
        """Test replacing localhost URL in kubeconfig."""
        kubeconfig = """apiVersion: v1
clusters:
- cluster:
    server: https://localhost:6443
  name: envtest
"""
        new_url = "https://192.168.1.100:32768"
        result = re.sub(
            r"server: https://(?:localhost|127\.0\.0\.1):\d+",
            f"server: {new_url}",
            kubeconfig,
        )

        assert new_url in result
        assert "localhost" not in result

    def test_replace_127_0_0_1_url(self) -> None:
        """Test replacing 127.0.0.1 URL in kubeconfig."""
        kubeconfig = """apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: envtest
"""
        new_url = "https://host.docker.internal:45678"
        result = re.sub(
            r"server: https://(?:localhost|127\.0\.0\.1):\d+",
            f"server: {new_url}",
            kubeconfig,
        )

        assert new_url in result
        assert "127.0.0.1" not in result

    def test_no_replacement_for_external_url(self) -> None:
        """Test that external URLs are not replaced."""
        kubeconfig = """apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.default:443
  name: envtest
"""
        new_url = "https://192.168.1.100:32768"
        result = re.sub(
            r"server: https://(?:localhost|127\.0\.0\.1):\d+",
            f"server: {new_url}",
            kubeconfig,
        )

        # Original URL should remain unchanged
        assert "kubernetes.default:443" in result
        assert new_url not in result


class TestModuleExports:
    """Tests for module exports."""

    def test_envtest_container_exported(self) -> None:
        """Test that EnvtestContainer is exported from the package."""
        from testcontainers_envtest import EnvtestContainer

        assert EnvtestContainer is not None

    def test_version_exported(self) -> None:
        """Test that __version__ is exported from the package."""
        from testcontainers_envtest import __version__

        assert __version__ is not None
        assert isinstance(__version__, str)
        # Should be a valid semver-like version
        assert re.match(r"^\d+\.\d+\.\d+", __version__)

    def test_all_exports(self) -> None:
        """Test that __all__ contains expected exports."""
        import testcontainers_envtest

        assert hasattr(testcontainers_envtest, "__all__")
        assert "EnvtestContainer" in testcontainers_envtest.__all__
