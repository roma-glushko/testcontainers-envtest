# testcontainers-envtest

[![Docker](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/docker.yml/badge.svg)](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/docker.yml)
[![Go](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/go.yml/badge.svg)](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/go.yml)
[![Python](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/python.yml/badge.svg)](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/python.yml)
[![Java](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/java.yml/badge.svg)](https://github.com/roma-glushko/testcontainers-envtest/actions/workflows/java.yml)

A [Testcontainers](https://testcontainers.org/) integration for [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) - a lightweight Kubernetes API server and etcd for testing Kubernetes controllers and operators.

## Why Envtest?

| Feature        | Envtest                        | K3s                            |
|----------------|--------------------------------|--------------------------------|
| Components     | kube-apiserver + etcd only     | Full Kubernetes distribution   |
| Startup time   | ~3 seconds                     | ~8 seconds                     |
| Use case       | Controller/operator unit tests | Full cluster integration tests |
| Resource usage | Minimal                        | Higher                         |

Envtest provides a faster and lighter alternative to full Kubernetes clusters (like K3s) for testing controllers and operators. It only runs the components you need for controller testing: the Kubernetes API server and etcd.

## Supported Languages

- **Go** - `github.com/roma-glushko/testcontainers-envtest/go`
- **Python** - `testcontainers-envtest` on PyPI
- **Java** - `io.github.roma-glushko:testcontainers-envtest` on Maven Central

## Supported Kubernetes Versions

- 1.34.x
- 1.33.x
- 1.32.x
- 1.31.x

## Installation

### Go

```bash
go get github.com/roma-glushko/testcontainers-envtest/go
```

### Python

```bash
# Using uv (recommended)
uv add testcontainers-envtest

# Using pip
pip install testcontainers-envtest
```

### Java

```xml
<dependency>
    <groupId>io.github.roma-glushko</groupId>
    <artifactId>testcontainers-envtest</artifactId>
    <version>0.1.0</version>
    <scope>test</scope>
</dependency>
```

## Usage

### Go

```go
package mycontroller_test

import (
    "context"
    "testing"

	"github.com/stretchr/testify/require"
    "github.com/roma-glushko/testcontainers-envtest/go"
    "github.com/testcontainers/testcontainers-go"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMyController(t *testing.T) {
    ctx := t.Context()

    // Start an envtest container
    k8s, err := envtest.Run(ctx)
	require.NoError(t, err)
    
    defer testcontainers.TerminateContainer(k8s)

    // Get a REST config for the Kubernetes client
    cfg, err := k8s.RESTConfig(ctx)
	require.NoError(t, err)

    // Create a Kubernetes clientset
    clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

    // Use the client in your tests
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	
    t.Logf("Found %d namespaces", len(namespaces.Items))
}
```

#### With a specific Kubernetes version

```go
container, err := envtest.Run(ctx,
    envtest.WithKubernetesVersion("1.34.0"),
)
```

### Python

```python
import pytest
from testcontainers_envtest import EnvtestContainer
from kubernetes import client

@pytest.fixture(scope="module")
def envtest():
    with EnvtestContainer() as container:
        yield container

def test_my_controller(envtest):
    # Get a configured kubernetes client
    api_client = envtest.get_kubernetes_client()

    # Use the client
    v1 = client.CoreV1Api(api_client)
    namespaces = v1.list_namespace()
    print(f"Found {len(namespaces.items)} namespaces")

    # Create test resources
    namespace = client.V1Namespace(
        metadata=client.V1ObjectMeta(name="test-ns")
    )
    v1.create_namespace(body=namespace)
```

#### With a specific Kubernetes version

```python
with EnvtestContainer(kubernetes_version="1.34.0") as envtest:
    # ...
```

### Java

```java
import io.github.romaglushko.testcontainers.envtest.EnvtestContainer;
import io.fabric8.kubernetes.client.KubernetesClient;
import org.junit.jupiter.api.Test;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

@Testcontainers
class MyControllerTest {

    @Container
    static EnvtestContainer envtest = new EnvtestContainer();

    @Test
    void testController() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            // Use the client
            var namespaces = client.namespaces().list();
            System.out.println("Found " + namespaces.getItems().size() + " namespaces");

            // Create test resources
            client.namespaces()
                .resource(new NamespaceBuilder()
                    .withNewMetadata().withName("test-ns").endMetadata()
                    .build())
                .create();
        }
    }
}
```

#### With a specific Kubernetes version

```java
@Container
static EnvtestContainer envtest = new EnvtestContainer("1.30.0");
```

## Docker Image

The Docker image is available on GitHub Container Registry:

```bash
docker pull ghcr.io/roma-glushko/testcontainers-envtest:latest

# Or a specific Kubernetes version
docker pull ghcr.io/roma-glushko/testcontainers-envtest:v1.34.0
```

### Running standalone

```bash
docker run -p 6443:6443 ghcr.io/roma-glushko/testcontainers-envtest:latest
```

The container will:
1. Start etcd
2. Start kube-apiserver
3. Generate certificates and kubeconfig
4. Expose the API server on port 6443

## License

Apache 2.0 License - see [LICENSE](LICENSE) for details.
