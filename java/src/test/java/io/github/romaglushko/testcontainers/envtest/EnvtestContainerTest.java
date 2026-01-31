package io.github.romaglushko.testcontainers.envtest;

import io.fabric8.kubernetes.api.model.ConfigMap;
import io.fabric8.kubernetes.api.model.ConfigMapBuilder;
import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceBuilder;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import org.junit.jupiter.api.Test;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.testcontainers.utility.DockerImageName;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Tests for {@link EnvtestContainer}.
 *
 * <p>The ENVTEST_IMAGE environment variable can be set to test against a custom image.</p>
 */
@Testcontainers
class EnvtestContainerTest {

    @Container
    static EnvtestContainer envtest = createEnvtestContainer();

    /**
     * Creates an EnvtestContainer, optionally using a custom image from ENVTEST_IMAGE env var.
     */
    private static EnvtestContainer createEnvtestContainer() {
        String customImage = System.getenv("ENVTEST_IMAGE");
        if (customImage != null && !customImage.isEmpty()) {
            return new EnvtestContainer(DockerImageName.parse(customImage));
        }
        return new EnvtestContainer();
    }

    @Test
    void shouldReturnApiServerUrl() {
        String url = envtest.getApiServerUrl();

        assertThat(url).isNotNull();
        assertThat(url).startsWith("https://");
        assertThat(url).containsPattern(":\\d+$"); // Should have a port
    }

    @Test
    void shouldReturnKubernetesVersion() {
        String version = envtest.getKubernetesVersion();

        // Version should be non-empty and match pattern like "1.35.0"
        assertThat(version).isNotEmpty();
        assertThat(version).matches("\\d+\\.\\d+\\.\\d+");
    }

    @Test
    void shouldReturnValidKubeconfig() {
        String kubeconfig = envtest.getKubeconfig();

        assertThat(kubeconfig).isNotNull();
        assertThat(kubeconfig).contains("apiVersion: v1");
        assertThat(kubeconfig).contains("kind: Config");
        assertThat(kubeconfig).contains("clusters:");
        assertThat(kubeconfig).contains("users:");
        assertThat(kubeconfig).contains("contexts:");
    }

    @Test
    void shouldReturnKubeconfigWithCorrectUrl() {
        String kubeconfig = envtest.getKubeconfig();
        String apiUrl = envtest.getApiServerUrl();

        assertThat(kubeconfig).contains(apiUrl);
    }

    @Test
    void shouldWriteKubeconfigToTempFile() throws Exception {
        Path path = envtest.getKubeconfigPath();

        assertThat(path).isNotNull();
        assertThat(Files.exists(path)).isTrue();

        String content = Files.readString(path);
        assertThat(content).contains("apiVersion: v1");
        assertThat(content).contains("kind: Config");

        // Clean up
        Files.deleteIfExists(path);
    }

    @Test
    void shouldProvideWorkingKubernetesClient() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            assertThat(client).isNotNull();

            NamespaceList namespaces = client.namespaces().list();
            assertThat(namespaces).isNotNull();
            assertThat(namespaces.getItems()).isNotEmpty();

            // Should have the default namespace
            boolean hasDefault = namespaces.getItems().stream()
                .anyMatch(ns -> "default".equals(ns.getMetadata().getName()));
            assertThat(hasDefault).isTrue();
        }
    }

    @Test
    void shouldAllowCreatingResources() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            // Create a test namespace
            Namespace namespace = new NamespaceBuilder()
                .withNewMetadata()
                .withName("test-namespace")
                .endMetadata()
                .build();

            Namespace created = client.namespaces().resource(namespace).create();
            assertThat(created.getMetadata().getName()).isEqualTo("test-namespace");

            // Verify the namespace exists
            Namespace got = client.namespaces().withName("test-namespace").get();
            assertThat(got).isNotNull();
            assertThat(got.getMetadata().getName()).isEqualTo("test-namespace");

            // Clean up
            client.namespaces().withName("test-namespace").delete();
        }
    }

    @Test
    void shouldCreateAndGetConfigMap() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            // Create a ConfigMap
            ConfigMap configMap = new ConfigMapBuilder()
                .withNewMetadata()
                .withName("test-configmap")
                .withNamespace("default")
                .endMetadata()
                .withData(Map.of("key1", "value1", "key2", "value2"))
                .build();

            ConfigMap created = client.configMaps()
                .inNamespace("default")
                .resource(configMap)
                .create();
            assertThat(created.getMetadata().getName()).isEqualTo("test-configmap");

            // Retrieve the ConfigMap
            ConfigMap got = client.configMaps()
                .inNamespace("default")
                .withName("test-configmap")
                .get();
            assertThat(got.getData()).containsEntry("key1", "value1");
            assertThat(got.getData()).containsEntry("key2", "value2");

            // Clean up
            client.configMaps().inNamespace("default").withName("test-configmap").delete();
        }
    }

    @Test
    void shouldCreateAndDeleteSecret() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            // Create a Secret
            Secret secret = new SecretBuilder()
                .withNewMetadata()
                .withName("test-secret")
                .withNamespace("default")
                .endMetadata()
                .withStringData(Map.of("password", "supersecret"))
                .build();

            Secret created = client.secrets()
                .inNamespace("default")
                .resource(secret)
                .create();
            assertThat(created.getMetadata().getName()).isEqualTo("test-secret");

            // Delete the Secret
            client.secrets().inNamespace("default").withName("test-secret").delete();

            // Verify it's deleted
            Secret got = client.secrets()
                .inNamespace("default")
                .withName("test-secret")
                .get();
            assertThat(got).isNull();
        }
    }

    @Test
    void shouldCreateService() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            // Create a Service
            Service service = new ServiceBuilder()
                .withNewMetadata()
                .withName("test-service")
                .withNamespace("default")
                .endMetadata()
                .withNewSpec()
                .withSelector(Map.of("app", "test"))
                .addNewPort()
                .withPort(80)
                .withTargetPort(new io.fabric8.kubernetes.api.model.IntOrString(8080))
                .endPort()
                .endSpec()
                .build();

            Service created = client.services()
                .inNamespace("default")
                .resource(service)
                .create();
            assertThat(created.getMetadata().getName()).isEqualTo("test-service");

            // Retrieve the Service
            Service got = client.services()
                .inNamespace("default")
                .withName("test-service")
                .get();
            assertThat(got.getSpec().getPorts()).hasSize(1);
            assertThat(got.getSpec().getPorts().get(0).getPort()).isEqualTo(80);

            // Clean up
            client.services().inNamespace("default").withName("test-service").delete();
        }
    }

    @Test
    void shouldListDefaultNamespaces() {
        try (KubernetesClient client = envtest.getKubernetesClient()) {
            NamespaceList namespaces = client.namespaces().list();

            // Envtest typically creates: default, kube-system, kube-public, kube-node-lease
            assertThat(namespaces.getItems()).hasSizeGreaterThanOrEqualTo(1);

            var namespaceNames = namespaces.getItems().stream()
                .map(ns -> ns.getMetadata().getName())
                .toList();
            assertThat(namespaceNames).contains("default");
        }
    }
}

/**
 * Tests for {@link EnvtestContainer} with custom Kubernetes versions.
 */
class EnvtestContainerVersionTest {

    @Test
    void shouldUseDefaultKubernetesVersion() {
        EnvtestContainer container = new EnvtestContainer();
        assertThat(container.getKubernetesVersion())
            .isEqualTo(EnvtestContainer.DEFAULT_KUBERNETES_VERSION);
    }

    @Test
    void shouldAcceptCustomKubernetesVersion() {
        EnvtestContainer container = new EnvtestContainer("1.30.0");
        assertThat(container.getKubernetesVersion()).isEqualTo("1.30.0");
    }
}
