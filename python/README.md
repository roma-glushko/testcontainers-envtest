# testcontainers-envtest

Testcontainers module for [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) - a lightweight Kubernetes API server and etcd for testing Kubernetes controllers and operators.

## Installation

```bash
# Using uv (recommended)
uv add testcontainers-envtest

# With kubernetes client support
uv add testcontainers-envtest --optional kubernetes

# Using pip
pip install testcontainers-envtest
pip install testcontainers-envtest[kubernetes]
```

## Usage

### Basic usage with kubernetes client

```python
from testcontainers_envtest import EnvtestContainer
from kubernetes import client

with EnvtestContainer() as envtest:
    # Get a configured kubernetes client
    api_client = envtest.get_kubernetes_client()

    # Use the client
    v1 = client.CoreV1Api(api_client)
    namespaces = v1.list_namespace()
    print(f"Found {len(namespaces.items)} namespaces")
```

### With pytest

```python
import pytest
from testcontainers_envtest import EnvtestContainer
from kubernetes import client

@pytest.fixture(scope="module")
def envtest():
    with EnvtestContainer() as container:
        yield container

def test_my_controller(envtest):
    api_client = envtest.get_kubernetes_client()
    v1 = client.CoreV1Api(api_client)

    # Create test resources
    namespace = client.V1Namespace(
        metadata=client.V1ObjectMeta(name="test-ns")
    )
    v1.create_namespace(body=namespace)

    # ... run your tests
```

### Get kubeconfig for external tools

```python
from testcontainers_envtest import EnvtestContainer

with EnvtestContainer() as envtest:
    # Get kubeconfig as a string
    kubeconfig = envtest.get_kubeconfig()

    # Or get a file path
    kubeconfig_path = envtest.get_kubeconfig_path()

    # Use with external tools
    import subprocess
    subprocess.run(["kubectl", "--kubeconfig", kubeconfig_path, "get", "namespaces"])
```

### Specify Kubernetes version

```python
from testcontainers_envtest import EnvtestContainer

# Use a specific Kubernetes version
with EnvtestContainer(kubernetes_version="1.35.0") as envtest:
    print(f"Running Kubernetes {envtest.kubernetes_version}")
```

## Supported Kubernetes versions

- 1.27.x
- 1.28.x
- 1.29.x
- 1.30.x
- 1.31.x

## License

MIT
