#!/bin/bash
set -e

# Script to generate self-signed certificates for Windows code signing testing
# WARNING: These certificates are for TESTING ONLY and should NEVER be used in production!

echo "=================================================="
echo "Windows Code Signing Test Certificate Generator"
echo "=================================================="
echo ""
echo "⚠️  WARNING: This creates a SELF-SIGNED certificate"
echo "⚠️  For TESTING purposes only!"
echo "⚠️  Production releases should use SignPath.io"
echo ""

# Check for OpenSSL
if ! command -v openssl &> /dev/null; then
    echo "Error: OpenSSL is not installed."
    echo "Please install OpenSSL first:"
    echo "  macOS: brew install openssl"
    echo "  Ubuntu/Debian: sudo apt-get install openssl"
    echo "  Windows: Use Git Bash or WSL with OpenSSL"
    exit 1
fi

# Configuration
CERT_DIR=".certs"
CERT_NAME="test-cert"
CERT_PASSWORD="${WINDOWS_CERT_PASSWORD:-test-password}"
DAYS_VALID=365

# Create certificate directory
mkdir -p "$CERT_DIR"

echo "Creating test certificate..."
echo ""

# Generate private key
echo "1. Generating private key..."
openssl genrsa -out "$CERT_DIR/$CERT_NAME.key" 2048 2>/dev/null

# Generate certificate signing request
echo "2. Creating certificate signing request..."
openssl req -new \
    -key "$CERT_DIR/$CERT_NAME.key" \
    -out "$CERT_DIR/$CERT_NAME.csr" \
    -subj "/C=US/ST=Test/L=Test/O=GoPCA Test/CN=GoPCA Test Certificate" 2>/dev/null

# Generate self-signed certificate
echo "3. Generating self-signed certificate..."
openssl x509 -req \
    -days $DAYS_VALID \
    -in "$CERT_DIR/$CERT_NAME.csr" \
    -signkey "$CERT_DIR/$CERT_NAME.key" \
    -out "$CERT_DIR/$CERT_NAME.crt" 2>/dev/null

# Create PKCS#12 file for osslsigncode
echo "4. Creating PKCS#12 file for code signing..."
openssl pkcs12 -export \
    -out "$CERT_DIR/$CERT_NAME.p12" \
    -inkey "$CERT_DIR/$CERT_NAME.key" \
    -in "$CERT_DIR/$CERT_NAME.crt" \
    -password "pass:$CERT_PASSWORD" 2>/dev/null

# Clean up intermediate files
rm -f "$CERT_DIR/$CERT_NAME.csr"

echo ""
echo "✅ Test certificate created successfully!"
echo ""
echo "Certificate details:"
echo "  Location: $CERT_DIR/$CERT_NAME.p12"
echo "  Password: $CERT_PASSWORD"
echo "  Valid for: $DAYS_VALID days"
echo ""
echo "To use this certificate:"
echo ""
echo "1. Create a .env file with:"
echo "   WINDOWS_CERT_FILE=$CERT_DIR/$CERT_NAME.p12"
echo "   WINDOWS_CERT_PASSWORD=$CERT_PASSWORD"
echo ""
echo "2. Run: make sign-windows"
echo ""
echo "⚠️  Remember: This certificate will trigger Windows security warnings!"
echo "⚠️  Only use for testing - never distribute binaries signed with this!"
echo ""