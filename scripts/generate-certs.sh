#!/bin/bash

# Heimdall TLS Certificate Generator
# This script generates self-signed certificates for development/testing purposes
# DO NOT use these certificates in production!

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/certs"
CERT_NAME="heimdall"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Heimdall TLS Certificate Generator${NC}"
echo "=================================="

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}Error: OpenSSL is not installed or not in PATH${NC}"
    echo "Please install OpenSSL to generate certificates"
    exit 1
fi

# Certificate configuration
cat > "${CERT_NAME}.conf" << EOF
[req]
default_bits = 2048
default_md = sha256
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = California
L = San Francisco
O = Heimdall Gateway
OU = Development
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

echo -e "${YELLOW}Generating private key...${NC}"
openssl genrsa -out "${CERT_NAME}.key" 2048

echo -e "${YELLOW}Generating certificate signing request...${NC}"
openssl req -new -key "${CERT_NAME}.key" -out "${CERT_NAME}.csr" -config "${CERT_NAME}.conf"

echo -e "${YELLOW}Generating self-signed certificate...${NC}"
openssl x509 -req -in "${CERT_NAME}.csr" -signkey "${CERT_NAME}.key" -out "${CERT_NAME}.crt" \
    -days 365 -extensions v3_req -extfile "${CERT_NAME}.conf"

# Clean up CSR and config files
rm "${CERT_NAME}.csr" "${CERT_NAME}.conf"

echo -e "${GREEN}Certificate generation completed!${NC}"
echo ""
echo "Generated files:"
echo "  - ${CERT_DIR}/${CERT_NAME}.crt (certificate)"
echo "  - ${CERT_DIR}/${CERT_NAME}.key (private key)"
echo ""
echo -e "${YELLOW}WARNING: These are self-signed certificates for development only!${NC}"
echo -e "${YELLOW}         DO NOT use them in production!${NC}"
echo ""
echo "To use with Heimdall, set these environment variables:"
echo "  HEIMDALL_TLS_CERT=${CERT_DIR}/${CERT_NAME}.crt"
echo "  HEIMDALL_TLS_KEY=${CERT_DIR}/${CERT_NAME}.key"
echo ""
echo "Or copy them to your system certificate location:"
echo "  sudo cp ${CERT_NAME}.crt /etc/ssl/certs/"
echo "  sudo cp ${CERT_NAME}.key /etc/ssl/private/"
echo ""
echo "To verify the certificate:"
echo "  openssl x509 -in ${CERT_NAME}.crt -text -noout"