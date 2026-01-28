package io.github.romaglushko.testcontainers.envtest;

import org.junit.jupiter.api.Test;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Unit tests for {@link EnvtestContainer} that don't require Docker.
 */
class EnvtestContainerUnitTest {

    @Test
    void shouldHaveCorrectDefaultImage() {
        assertThat(EnvtestContainer.DEFAULT_IMAGE)
            .isEqualTo("ghcr.io/roma-glushko/testcontainers-envtest");
    }

    @Test
    void shouldHaveCorrectDefaultTag() {
        assertThat(EnvtestContainer.DEFAULT_TAG)
            .isEqualTo("latest");
    }

    @Test
    void shouldHaveCorrectDefaultKubernetesVersion() {
        assertThat(EnvtestContainer.DEFAULT_KUBERNETES_VERSION)
            .isEqualTo("1.31.0");
    }

    @Test
    void shouldHaveCorrectApiServerPort() {
        assertThat(EnvtestContainer.API_SERVER_PORT)
            .isEqualTo(6443);
    }

    @Test
    void shouldHaveCorrectKubeconfigPath() {
        assertThat(EnvtestContainer.KUBECONFIG_PATH)
            .isEqualTo("/tmp/kubeconfig");
    }

    @Test
    void shouldUseDefaultKubernetesVersion() {
        EnvtestContainer container = new EnvtestContainer();
        assertThat(container.getKubernetesVersion())
            .isEqualTo(EnvtestContainer.DEFAULT_KUBERNETES_VERSION);
    }

    @Test
    void shouldAcceptCustomKubernetesVersion() {
        EnvtestContainer container = new EnvtestContainer("1.30.0");
        assertThat(container.getKubernetesVersion())
            .isEqualTo("1.30.0");
    }

    @Test
    void shouldExposeApiServerPort() {
        EnvtestContainer container = new EnvtestContainer();
        assertThat(container.getExposedPorts())
            .contains(EnvtestContainer.API_SERVER_PORT);
    }
}

/**
 * Tests for kubeconfig URL replacement logic.
 */
class KubeconfigParsingTest {

    private static final Pattern SERVER_URL_PATTERN =
        Pattern.compile("server: https://(?:localhost|127\\.0\\.0\\.1):\\d+");

    @Test
    void shouldReplaceLocalhostUrl() {
        String kubeconfig = """
            apiVersion: v1
            clusters:
            - cluster:
                server: https://localhost:6443
              name: envtest
            """;

        String newUrl = "https://192.168.1.100:32768";
        Matcher matcher = SERVER_URL_PATTERN.matcher(kubeconfig);
        String result = matcher.replaceAll("server: " + newUrl);

        assertThat(result).contains(newUrl);
        assertThat(result).doesNotContain("localhost:6443");
    }

    @Test
    void shouldReplace127001Url() {
        String kubeconfig = """
            apiVersion: v1
            clusters:
            - cluster:
                server: https://127.0.0.1:6443
              name: envtest
            """;

        String newUrl = "https://host.docker.internal:45678";
        Matcher matcher = SERVER_URL_PATTERN.matcher(kubeconfig);
        String result = matcher.replaceAll("server: " + newUrl);

        assertThat(result).contains(newUrl);
        assertThat(result).doesNotContain("127.0.0.1:6443");
    }

    @Test
    void shouldNotReplaceExternalUrl() {
        String kubeconfig = """
            apiVersion: v1
            clusters:
            - cluster:
                server: https://kubernetes.default:443
              name: envtest
            """;

        String newUrl = "https://192.168.1.100:32768";
        Matcher matcher = SERVER_URL_PATTERN.matcher(kubeconfig);
        String result = matcher.replaceAll("server: " + newUrl);

        // Original URL should remain unchanged
        assertThat(result).contains("kubernetes.default:443");
        assertThat(result).doesNotContain(newUrl);
    }

    @Test
    void shouldHandleEmptyKubeconfig() {
        String kubeconfig = "";
        String newUrl = "https://192.168.1.100:32768";
        Matcher matcher = SERVER_URL_PATTERN.matcher(kubeconfig);
        String result = matcher.replaceAll("server: " + newUrl);

        assertThat(result).isEmpty();
    }

    @Test
    void shouldHandleMultiplePorts() {
        String kubeconfig = """
            apiVersion: v1
            clusters:
            - cluster:
                server: https://localhost:32768
              name: envtest
            """;

        String newUrl = "https://192.168.1.100:45678";
        Matcher matcher = SERVER_URL_PATTERN.matcher(kubeconfig);
        String result = matcher.replaceAll("server: " + newUrl);

        assertThat(result).contains(newUrl);
        assertThat(result).doesNotContain("localhost:32768");
    }
}

/**
 * Tests for container configuration.
 */
class ContainerConfigurationTest {

    @Test
    void shouldBeGenericContainer() {
        EnvtestContainer container = new EnvtestContainer();
        assertThat(container).isInstanceOf(org.testcontainers.containers.GenericContainer.class);
    }

    @Test
    void shouldReturnSelfFromFluentMethods() {
        EnvtestContainer container = new EnvtestContainer();

        // Fluent methods should return the container itself
        EnvtestContainer result = container.withEnv("TEST", "value");
        assertThat(result).isSameAs(container);
    }
}
