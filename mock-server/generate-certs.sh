#!/bin/bash
set -e

CERTS_DIR="certs"
mkdir -p $CERTS_DIR

echo "Generating self-signed certificates for Mock DAO Server..."

# Generate DAO server private key and self-signed certificate
openssl genrsa -out $CERTS_DIR/dao-server.key 4096

openssl req -new -x509 -key $CERTS_DIR/dao-server.key -sha256 -subj "/C=HK/ST=HK/O=TEENet/CN=localhost" -days 365 -out $CERTS_DIR/dao-server.crt -extensions v3_req -config <(
cat <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = dao-server
DNS.3 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# Generate App Node private key and self-signed certificate
openssl genrsa -out $CERTS_DIR/app-node.key 4096

openssl req -new -x509 -key $CERTS_DIR/app-node.key -sha256 -subj "/C=HK/ST=HK/O=TEENet/CN=localhost" -days 365 -out $CERTS_DIR/app-node.crt -extensions v3_req -config <(
cat <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = app-node
DNS.3 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# Generate client private key and self-signed certificate
openssl genrsa -out $CERTS_DIR/client.key 4096

openssl req -new -x509 -key $CERTS_DIR/client.key -sha256 -subj "/C=HK/ST=HK/O=TEENet/CN=localhost" -days 365 -out $CERTS_DIR/client.crt -extensions v3_req -config <(
cat <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = test-client
DNS.3 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

echo "Self-signed certificates generated successfully:"
echo "  - DAO Server Certificate: $CERTS_DIR/dao-server.crt"
echo "  - DAO Server Private Key: $CERTS_DIR/dao-server.key"
echo "  - App Node Certificate: $CERTS_DIR/app-node.crt"
echo "  - App Node Private Key: $CERTS_DIR/app-node.key"
echo "  - Client Certificate: $CERTS_DIR/client.crt"
echo "  - Client Private Key: $CERTS_DIR/client.key"