"""Integration tests for the EnvtestContainer (require Docker)."""

import os

import pytest

from testcontainers_envtest import EnvtestContainer


@pytest.fixture(scope="module")
def envtest_container():
    """Create a shared envtest container for all tests in this module.

    If ENVTEST_IMAGE environment variable is set, uses that image for testing.
    """
    image = os.environ.get("ENVTEST_IMAGE")
    kwargs = {"image": image} if image else {}
    with EnvtestContainer(**kwargs) as container:
        yield container


@pytest.mark.integration
class TestEnvtestContainer:
    """Integration tests for EnvtestContainer functionality."""

    def test_get_api_server_url(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can get the API server URL."""
        url = envtest_container.get_api_server_url()

        assert url is not None
        assert url.startswith("https://")
        assert ":" in url  # Should have a port

    def test_get_kubeconfig(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can get the kubeconfig."""
        kubeconfig = envtest_container.get_kubeconfig()

        assert kubeconfig is not None
        assert len(kubeconfig) > 100  # Should be a reasonable size
        assert "apiVersion: v1" in kubeconfig
        assert "kind: Config" in kubeconfig
        assert "clusters:" in kubeconfig
        assert "users:" in kubeconfig
        assert "contexts:" in kubeconfig

    def test_get_kubeconfig_has_correct_url(
        self, envtest_container: EnvtestContainer
    ) -> None:
        """Test that the kubeconfig has the correct external URL."""
        kubeconfig = envtest_container.get_kubeconfig()
        api_url = envtest_container.get_api_server_url()

        # The kubeconfig should contain the external API URL
        assert api_url in kubeconfig

    def test_kubernetes_version(self, envtest_container: EnvtestContainer) -> None:
        """Test that the kubernetes version is returned correctly."""
        version = envtest_container.kubernetes_version

        assert version is not None
        # Version should match the expected pattern (e.g., 1.35.0)
        import re
        assert re.match(r"^\d+\.\d+\.\d+$", version), f"Invalid version format: {version}"

    def test_get_kubernetes_client(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can get a working Kubernetes client."""
        from kubernetes import client

        api_client = envtest_container.get_kubernetes_client()

        assert api_client is not None

        # Use the client to list namespaces
        v1 = client.CoreV1Api(api_client)
        namespaces = v1.list_namespace()

        assert namespaces is not None
        assert len(namespaces.items) > 0

        # Check for default namespace
        namespace_names = [ns.metadata.name for ns in namespaces.items]
        assert "default" in namespace_names

    def test_create_namespace(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can create a namespace using the client."""
        from kubernetes import client

        api_client = envtest_container.get_kubernetes_client()
        v1 = client.CoreV1Api(api_client)

        # Create a test namespace
        namespace = client.V1Namespace(
            metadata=client.V1ObjectMeta(name="test-namespace")
        )

        created = v1.create_namespace(body=namespace)
        assert created.metadata.name == "test-namespace"

        # Verify the namespace exists
        got = v1.read_namespace(name="test-namespace")
        assert got.metadata.name == "test-namespace"

        # Clean up
        v1.delete_namespace(name="test-namespace")

    def test_create_and_get_configmap(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can create and retrieve a ConfigMap."""
        from kubernetes import client

        api_client = envtest_container.get_kubernetes_client()
        v1 = client.CoreV1Api(api_client)

        # Create a ConfigMap
        configmap = client.V1ConfigMap(
            metadata=client.V1ObjectMeta(name="test-configmap", namespace="default"),
            data={"key1": "value1", "key2": "value2"},
        )

        created = v1.create_namespaced_config_map(namespace="default", body=configmap)
        assert created.metadata.name == "test-configmap"

        # Retrieve the ConfigMap
        got = v1.read_namespaced_config_map(name="test-configmap", namespace="default")
        assert got.data["key1"] == "value1"
        assert got.data["key2"] == "value2"

        # Clean up
        v1.delete_namespaced_config_map(name="test-configmap", namespace="default")

    def test_create_and_delete_secret(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can create and delete a Secret."""
        from kubernetes import client

        api_client = envtest_container.get_kubernetes_client()
        v1 = client.CoreV1Api(api_client)

        # Create a Secret
        secret = client.V1Secret(
            metadata=client.V1ObjectMeta(name="test-secret", namespace="default"),
            string_data={"password": "supersecret"},
        )

        created = v1.create_namespaced_secret(namespace="default", body=secret)
        assert created.metadata.name == "test-secret"

        # Delete the Secret
        v1.delete_namespaced_secret(name="test-secret", namespace="default")

        # Verify it's deleted
        with pytest.raises(client.exceptions.ApiException) as exc_info:
            v1.read_namespaced_secret(name="test-secret", namespace="default")
        assert exc_info.value.status == 404


@pytest.mark.integration
class TestEnvtestContainerKubeconfigPath:
    """Integration tests for kubeconfig file path functionality."""

    def test_get_kubeconfig_path(self, envtest_container: EnvtestContainer) -> None:
        """Test that we can get a kubeconfig file path."""
        path = envtest_container.get_kubeconfig_path()

        assert path is not None
        assert os.path.exists(path)
        assert path.endswith(".kubeconfig")

        # Read and verify the content
        with open(path) as f:
            content = f.read()

        assert "apiVersion: v1" in content
        assert "kind: Config" in content

    def test_kubeconfig_path_is_readable(self, envtest_container: EnvtestContainer) -> None:
        """Test that the kubeconfig file is readable and valid."""
        path = envtest_container.get_kubeconfig_path()

        # File should be readable
        with open(path) as f:
            content = f.read()

        # Should be valid YAML (basic check)
        assert content.strip().startswith("apiVersion:")


class TestEnvtestContainerWithVersion:
    """Unit tests for EnvtestContainer with a specific Kubernetes version."""

    def test_custom_kubernetes_version(self) -> None:
        """Test that we can specify a custom Kubernetes version."""
        container = EnvtestContainer(kubernetes_version="1.34.1")

        assert container.kubernetes_version == "1.34.1"

    def test_default_kubernetes_version(self) -> None:
        """Test that the default Kubernetes version is set correctly."""
        container = EnvtestContainer()

        assert container.kubernetes_version == EnvtestContainer.DEFAULT_KUBERNETES_VERSION
