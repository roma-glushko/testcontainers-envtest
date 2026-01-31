#!/bin/bash
set -e

# Configuration
ENVTEST_BIN_DIR="/usr/local/bin/envtest"
ETCD_PORT="${ETCD_PORT:-2379}"
API_SERVER_PORT="${API_SERVER_PORT:-6443}"
KUBECONFIG_PATH="${KUBECONFIG_PATH:-/tmp/kubeconfig}"
DATA_DIR="/tmp/envtest"
CERTS_CONF_DIR="/etc/envtest/certs"

# Create data directory
mkdir -p "${DATA_DIR}"

# Find the correct binary directory (setup-envtest creates versioned subdirs)
BINARY_DIR=$(find ${ENVTEST_BIN_DIR} -name "kube-apiserver" -type f -exec dirname {} \; | head -1)

if [ -z "${BINARY_DIR}" ]; then
    echo "ERROR: Could not find envtest binaries"
    exit 1
fi

echo "Using envtest binaries from: ${BINARY_DIR}"

ETCD_BINARY="${BINARY_DIR}/etcd"
APISERVER_BINARY="${BINARY_DIR}/kube-apiserver"

# Verify binaries exist
if [ ! -x "${ETCD_BINARY}" ]; then
    echo "ERROR: etcd binary not found or not executable at ${ETCD_BINARY}"
    exit 1
fi

if [ ! -x "${APISERVER_BINARY}" ]; then
    echo "ERROR: kube-apiserver binary not found or not executable at ${APISERVER_BINARY}"
    exit 1
fi

# Generate certificates for the API server
echo "Generating certificates..."
mkdir -p "${DATA_DIR}/certs"

# Generate CA with key usage extension (required for Python 3.13+)
openssl genrsa -out "${DATA_DIR}/certs/ca.key" 2048 2>/dev/null
openssl req -x509 -new -nodes -key "${DATA_DIR}/certs/ca.key" \
    -subj "/CN=envtest-ca" \
    -days 365 -out "${DATA_DIR}/certs/ca.crt" \
    -config "${CERTS_CONF_DIR}/ca.conf" 2>/dev/null

# Generate API server certificate
openssl genrsa -out "${DATA_DIR}/certs/apiserver.key" 2048 2>/dev/null
openssl req -new -key "${DATA_DIR}/certs/apiserver.key" \
    -subj "/CN=kube-apiserver" \
    -out "${DATA_DIR}/certs/apiserver.csr" \
    -config "${CERTS_CONF_DIR}/apiserver.conf" 2>/dev/null

openssl x509 -req -in "${DATA_DIR}/certs/apiserver.csr" \
    -CA "${DATA_DIR}/certs/ca.crt" \
    -CAkey "${DATA_DIR}/certs/ca.key" \
    -CAcreateserial \
    -out "${DATA_DIR}/certs/apiserver.crt" \
    -days 365 \
    -extensions v3_req \
    -extfile "${CERTS_CONF_DIR}/apiserver.conf" 2>/dev/null

# Generate client certificate for kubeconfig
openssl genrsa -out "${DATA_DIR}/certs/client.key" 2048 2>/dev/null
openssl req -new -key "${DATA_DIR}/certs/client.key" \
    -subj "/CN=admin/O=system:masters" \
    -out "${DATA_DIR}/certs/client.csr" \
    -config "${CERTS_CONF_DIR}/client.conf" 2>/dev/null
openssl x509 -req -in "${DATA_DIR}/certs/client.csr" \
    -CA "${DATA_DIR}/certs/ca.crt" \
    -CAkey "${DATA_DIR}/certs/ca.key" \
    -CAcreateserial \
    -out "${DATA_DIR}/certs/client.crt" \
    -days 365 \
    -extensions v3_req \
    -extfile "${CERTS_CONF_DIR}/client.conf" 2>/dev/null

echo "Certificates generated successfully"

# Start etcd in the background
echo "Starting etcd on port ${ETCD_PORT}..."
"${ETCD_BINARY}" \
    --data-dir="${DATA_DIR}/etcd" \
    --listen-client-urls="http://127.0.0.1:${ETCD_PORT}" \
    --advertise-client-urls="http://127.0.0.1:${ETCD_PORT}" \
    --listen-peer-urls="http://127.0.0.1:2380" \
    --initial-advertise-peer-urls="http://127.0.0.1:2380" \
    --initial-cluster="default=http://127.0.0.1:2380" \
    --log-level=error \
    &

ETCD_PID=$!

# Wait for etcd to be ready
echo "Waiting for etcd to be ready..."
for i in {1..30}; do
    if curl -s "http://127.0.0.1:${ETCD_PORT}/health" | grep -q "true"; then
        echo "etcd is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERROR: etcd failed to start"
        exit 1
    fi
    sleep 1
done

# Start kube-apiserver
echo "Starting kube-apiserver on port ${API_SERVER_PORT}..."
"${APISERVER_BINARY}" \
    --etcd-servers="http://127.0.0.1:${ETCD_PORT}" \
    --bind-address=0.0.0.0 \
    --secure-port="${API_SERVER_PORT}" \
    --tls-cert-file="${DATA_DIR}/certs/apiserver.crt" \
    --tls-private-key-file="${DATA_DIR}/certs/apiserver.key" \
    --client-ca-file="${DATA_DIR}/certs/ca.crt" \
    --service-account-key-file="${DATA_DIR}/certs/apiserver.key" \
    --service-account-signing-key-file="${DATA_DIR}/certs/apiserver.key" \
    --service-account-issuer="https://kubernetes.default.svc" \
    --authorization-mode=RBAC \
    --allow-privileged=true \
    --disable-admission-plugins=ServiceAccount \
    --service-cluster-ip-range=10.0.0.0/24 \
    --v=0 \
    &

APISERVER_PID=$!

# Wait for API server to be ready
echo "Waiting for kube-apiserver to be ready..."
for i in {1..60}; do
    if curl -sk "https://localhost:${API_SERVER_PORT}/healthz" | grep -q "ok"; then
        echo "kube-apiserver is ready"
        break
    fi
    if [ $i -eq 60 ]; then
        echo "ERROR: kube-apiserver failed to start"
        exit 1
    fi
    sleep 1
done

# Generate kubeconfig
echo "Generating kubeconfig at ${KUBECONFIG_PATH}..."
CA_DATA=$(base64 -w 0 "${DATA_DIR}/certs/ca.crt" 2>/dev/null || base64 "${DATA_DIR}/certs/ca.crt")
CLIENT_CERT_DATA=$(base64 -w 0 "${DATA_DIR}/certs/client.crt" 2>/dev/null || base64 "${DATA_DIR}/certs/client.crt")
CLIENT_KEY_DATA=$(base64 -w 0 "${DATA_DIR}/certs/client.key" 2>/dev/null || base64 "${DATA_DIR}/certs/client.key")

cat > "${KUBECONFIG_PATH}" <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: ${CA_DATA}
    server: https://localhost:${API_SERVER_PORT}
  name: envtest
contexts:
- context:
    cluster: envtest
    user: admin
  name: envtest
current-context: envtest
users:
- name: admin
  user:
    client-certificate-data: ${CLIENT_CERT_DATA}
    client-key-data: ${CLIENT_KEY_DATA}
EOF

echo "Kubeconfig generated successfully"

# Also write the CA cert, client cert, and client key to separate files for easy access
cp "${DATA_DIR}/certs/ca.crt" /tmp/ca.crt
cp "${DATA_DIR}/certs/client.crt" /tmp/client.crt
cp "${DATA_DIR}/certs/client.key" /tmp/client.key

echo ""
echo "============================================"
echo "Envtest is ready!"
echo "API Server: https://localhost:${API_SERVER_PORT}"
echo "Kubeconfig: ${KUBECONFIG_PATH}"
echo "============================================"
echo ""

# Handle shutdown gracefully
trap 'echo "Shutting down..."; kill $APISERVER_PID $ETCD_PID 2>/dev/null; exit 0' SIGTERM SIGINT

# Keep the container running
wait $APISERVER_PID
