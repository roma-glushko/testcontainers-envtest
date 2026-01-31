"""Envtest container module for testcontainers."""

from __future__ import annotations

import re
import tempfile
from typing import TYPE_CHECKING

from testcontainers.core.container import DockerContainer
from testcontainers.core.waiting_utils import wait_for_logs

if TYPE_CHECKING:
    from kubernetes import client as k8s_client


class EnvtestContainer(DockerContainer):
    """
    Envtest container for testing Kubernetes controllers and operators.

    Envtest provides a lightweight Kubernetes API server and etcd for testing
    without requiring a full Kubernetes cluster. This is faster and lighter
    than alternatives like K3s.

    Example usage::

        from testcontainers_envtest import EnvtestContainer
        from kubernetes import client

        with EnvtestContainer() as envtest:
            k8s_client = envtest.get_kubernetes_client()
            v1 = client.CoreV1Api(k8s_client)
            namespaces = v1.list_namespace()
            print(f"Found {len(namespaces.items)} namespaces")

    Args:
        image: Docker image to use. Defaults to ghcr.io/roma-glushko/testcontainers-envtest:latest
        kubernetes_version: Kubernetes version to use. If specified, will use
            the appropriate image tag. Supported versions: 1.27, 1.28, 1.29, 1.30, 1.31
        **kwargs: Additional arguments passed to DockerContainer
    """

    DEFAULT_IMAGE = "ghcr.io/roma-glushko/testcontainers-envtest:latest"
    DEFAULT_KUBERNETES_VERSION = "1.31.0"
    API_SERVER_PORT = 6443
    KUBECONFIG_PATH = "/tmp/kubeconfig"

    def __init__(
        self,
        image: str | None = None,
        kubernetes_version: str | None = None,
        **kwargs: object,
    ) -> None:
        self._kubernetes_version = kubernetes_version or self.DEFAULT_KUBERNETES_VERSION

        # Determine the image to use
        if image is None:
            if kubernetes_version and kubernetes_version != self.DEFAULT_KUBERNETES_VERSION:
                image = f"ghcr.io/roma-glushko/testcontainers-envtest:v{kubernetes_version}"
            else:
                image = self.DEFAULT_IMAGE

        super().__init__(image=image, **kwargs)
        self.with_exposed_ports(self.API_SERVER_PORT)

    def start(self) -> EnvtestContainer:
        """Start the envtest container and wait for it to be ready."""
        super().start()
        # Wait for the container to be ready
        wait_for_logs(self, "Envtest is ready!", timeout=120)
        return self

    @property
    def kubernetes_version(self) -> str:
        """Return the Kubernetes version of this envtest container."""
        return self._kubernetes_version

    def get_api_server_url(self) -> str:
        """
        Get the URL of the Kubernetes API server.

        Returns:
            The API server URL (e.g., https://localhost:32768)
        """
        host = self.get_container_host_ip()
        port = self.get_exposed_port(self.API_SERVER_PORT)
        return f"https://{host}:{port}"

    def get_kubeconfig(self) -> str:
        """
        Get the kubeconfig YAML content for connecting to the API server.

        The kubeconfig is modified to use the correct external host and port.

        Returns:
            Kubeconfig YAML string
        """
        # Read the kubeconfig from the container
        exit_code, output = self.exec(f"cat {self.KUBECONFIG_PATH}")
        if exit_code != 0:
            raise RuntimeError(f"Failed to read kubeconfig: {output}")

        kubeconfig = output.decode("utf-8")

        # Replace the server URL with the external address
        api_server_url = self.get_api_server_url()
        kubeconfig = re.sub(
            r"server: https://(?:localhost|127\.0\.0\.1):\d+",
            f"server: {api_server_url}",
            kubeconfig,
        )

        return kubeconfig

    def get_kubeconfig_path(self) -> str:
        """
        Get the kubeconfig written to a temporary file.

        This is useful for tools that require a file path rather than the
        kubeconfig content directly.

        Returns:
            Path to a temporary file containing the kubeconfig
        """
        kubeconfig = self.get_kubeconfig()

        # Write to a temporary file
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".kubeconfig", delete=False
        ) as f:
            f.write(kubeconfig)
            return f.name

    def get_kubernetes_client(self) -> k8s_client.ApiClient:
        """
        Get a configured Kubernetes API client.

        Requires the 'kubernetes' package to be installed:
            pip install testcontainers-envtest[kubernetes]

        Returns:
            A configured kubernetes.client.ApiClient instance

        Raises:
            ImportError: If the kubernetes package is not installed
        """
        try:
            from kubernetes import client as k8s_client
            from kubernetes import config as k8s_config
        except ImportError as e:
            raise ImportError(
                "The 'kubernetes' package is required. "
                "Install it with: pip install testcontainers-envtest[kubernetes]"
            ) from e

        kubeconfig = self.get_kubeconfig()

        # Load the kubeconfig from a string
        loader = k8s_config.kube_config.KubeConfigLoader(
            config_dict=k8s_config.kube_config.yaml.safe_load(kubeconfig)
        )

        configuration = k8s_client.Configuration()
        loader.load_and_set(configuration)

        return k8s_client.ApiClient(configuration)
