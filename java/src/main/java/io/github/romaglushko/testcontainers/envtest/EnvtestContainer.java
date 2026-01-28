package io.github.romaglushko.testcontainers.envtest;

import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.wait.strategy.Wait;
import org.testcontainers.utility.DockerImageName;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Testcontainers module for envtest - a lightweight Kubernetes API server and etcd
 * for testing Kubernetes controllers and operators.
 *
 * <p>Envtest provides a faster and lighter alternative to full Kubernetes clusters
 * (like K3s) for controller unit tests. It only runs kube-apiserver and etcd,
 * without kubelet, controller-manager, or other components.</p>
 *
 * <h2>Usage Example:</h2>
 * <pre>{@code
 * @Testcontainers
 * class MyControllerTest {
 *     @Container
 *     static EnvtestContainer envtest = new EnvtestContainer();
 *
 *     @Test
 *     void testController() {
 *         try (KubernetesClient client = envtest.getKubernetesClient()) {
 *             NamespaceList namespaces = client.namespaces().list();
 *             assertThat(namespaces.getItems()).isNotEmpty();
 *         }
 *     }
 * }
 * }</pre>
 */
public class EnvtestContainer extends GenericContainer<EnvtestContainer> {

    /** Default Docker image for envtest */
    public static final String DEFAULT_IMAGE = "ghcr.io/roma-glushko/testcontainers-envtest";

    /** Default image tag (latest Kubernetes version) */
    public static final String DEFAULT_TAG = "latest";

    /** Default Kubernetes version */
    public static final String DEFAULT_KUBERNETES_VERSION = "1.31.0";

    /** API server port inside the container */
    public static final int API_SERVER_PORT = 6443;

    /** Path to kubeconfig inside the container */
    public static final String KUBECONFIG_PATH = "/tmp/kubeconfig";

    private final String kubernetesVersion;

    /**
     * Create a new EnvtestContainer with the default image and Kubernetes version.
     */
    public EnvtestContainer() {
        this(DEFAULT_KUBERNETES_VERSION);
    }

    /**
     * Create a new EnvtestContainer with a specific Kubernetes version.
     *
     * @param kubernetesVersion The Kubernetes version to use (e.g., "1.30.0")
     */
    public EnvtestContainer(String kubernetesVersion) {
        this(DockerImageName.parse(DEFAULT_IMAGE + ":v" + kubernetesVersion), kubernetesVersion);
    }

    /**
     * Create a new EnvtestContainer with a custom Docker image.
     *
     * @param dockerImageName The Docker image to use
     */
    public EnvtestContainer(DockerImageName dockerImageName) {
        this(dockerImageName, DEFAULT_KUBERNETES_VERSION);
    }

    private EnvtestContainer(DockerImageName dockerImageName, String kubernetesVersion) {
        super(dockerImageName);
        this.kubernetesVersion = kubernetesVersion;

        withExposedPorts(API_SERVER_PORT);
        waitingFor(Wait.forLogMessage(".*Envtest is ready!.*", 1));
    }

    /**
     * Get the Kubernetes version of this envtest container.
     *
     * @return The Kubernetes version
     */
    public String getKubernetesVersion() {
        return kubernetesVersion;
    }

    /**
     * Get the URL of the Kubernetes API server.
     *
     * @return The API server URL (e.g., https://localhost:32768)
     */
    public String getApiServerUrl() {
        return String.format("https://%s:%d", getHost(), getMappedPort(API_SERVER_PORT));
    }

    /**
     * Get the kubeconfig YAML content for connecting to the API server.
     *
     * <p>The kubeconfig is modified to use the correct external host and port.</p>
     *
     * @return Kubeconfig YAML string
     */
    public String getKubeconfig() {
        try {
            ExecResult result = execInContainer("cat", KUBECONFIG_PATH);
            if (result.getExitCode() != 0) {
                throw new RuntimeException("Failed to read kubeconfig: " + result.getStderr());
            }

            String kubeconfig = result.getStdout();

            // Replace the server URL with the external address
            String apiServerUrl = getApiServerUrl();
            Pattern pattern = Pattern.compile("server: https://(?:localhost|127\\.0\\.0\\.1):\\d+");
            Matcher matcher = pattern.matcher(kubeconfig);
            kubeconfig = matcher.replaceAll("server: " + apiServerUrl);

            return kubeconfig;
        } catch (IOException | InterruptedException e) {
            throw new RuntimeException("Failed to get kubeconfig", e);
        }
    }

    /**
     * Get the kubeconfig written to a temporary file.
     *
     * <p>This is useful for tools that require a file path rather than the
     * kubeconfig content directly.</p>
     *
     * @return Path to a temporary file containing the kubeconfig
     */
    public Path getKubeconfigPath() {
        try {
            Path tempFile = Files.createTempFile("envtest-kubeconfig-", ".yaml");
            Files.writeString(tempFile, getKubeconfig(), StandardCharsets.UTF_8);
            return tempFile;
        } catch (IOException e) {
            throw new RuntimeException("Failed to write kubeconfig to temp file", e);
        }
    }

    /**
     * Get a configured Fabric8 Kubernetes client.
     *
     * <p>Requires the fabric8 kubernetes-client dependency to be on the classpath.</p>
     *
     * @return A configured KubernetesClient instance
     * @throws RuntimeException if the fabric8 client is not available
     */
    public io.fabric8.kubernetes.client.KubernetesClient getKubernetesClient() {
        try {
            Class.forName("io.fabric8.kubernetes.client.KubernetesClient");
        } catch (ClassNotFoundException e) {
            throw new RuntimeException(
                "Fabric8 Kubernetes client is not on the classpath. " +
                "Add the dependency: io.fabric8:kubernetes-client"
            );
        }

        String kubeconfig = getKubeconfig();
        io.fabric8.kubernetes.client.Config config =
            io.fabric8.kubernetes.client.Config.fromKubeconfig(kubeconfig);

        return new io.fabric8.kubernetes.client.KubernetesClientBuilder()
            .withConfig(config)
            .build();
    }
}
