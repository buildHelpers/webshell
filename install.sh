#!/bin/bash

# WebShell Install Script
set -e

# Configuration
REPO="buildHelpers/webshell"
VERSION=${1:-"latest"}
BINARY_NAME="chumen-webshell"
SERVICE_NAME="chumen-webshell"
INSTALL_SERVICE=false

if [ "${1:-}" = "--service" ]; then
    INSTALL_SERVICE=true
fi

echo "WebShell Install Script"
echo "Repository: ${REPO}"
echo "Version: ${VERSION}"
echo ""

# Detect platform
OS=""
ARCH=""

case "$(uname -s)" in
    Linux*)     OS="linux";;
    Darwin*)    OS="darwin";;
    CYGWIN*)    OS="windows";;
    MINGW*)     OS="windows";;
    MSYS*)      OS="windows";;
    *)          echo "Unsupported operating system"; exit 1;;
esac

case "$(uname -m)" in
    x86_64)     ARCH="amd64";;
    amd64)      ARCH="amd64";;
    arm64)      ARCH="arm64";;
    aarch64)    ARCH="arm64";;
    *)          echo "Unsupported architecture: $(uname -m)"; exit 1;;
esac

PLATFORM="${OS}_${ARCH}"
echo "Detected: ${OS} ${ARCH}"

# Get version
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi
VERSION=${VERSION#v}

# Set filename
FILENAME="webshell_${PLATFORM}"
if [ "$PLATFORM" = "windows_amd64" ]; then
    FILENAME="webshell.exe"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

echo "Downloading WebShell v${VERSION} for ${PLATFORM}"
echo "URL: ${DOWNLOAD_URL}"

# Download
if curl -L -o "${FILENAME}" "${DOWNLOAD_URL}"; then
    echo "Download successful"
else
    echo "Download failed"
    echo "Make sure the release exists at: https://github.com/${REPO}/releases/tag/v${VERSION}"
    exit 1
fi

# Install
echo "Installing WebShell"
chmod +x "${FILENAME}"

if [ -w "/usr/local/bin" ] || sudo -n true 2>/dev/null; then
    sudo mv "${FILENAME}" "/usr/local/bin/${BINARY_NAME}"
    echo "Installed to /usr/local/bin/${BINARY_NAME}"
else
    mkdir -p "$HOME/.local/bin"
    mv "${FILENAME}" "$HOME/.local/bin/${BINARY_NAME}"
    echo "Installed to $HOME/.local/bin/${BINARY_NAME}"
    echo "Add $HOME/.local/bin to your PATH if not already added"
    echo "Add this line to your shell profile:"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "Alternative: you can run ~/.local/bin/chumen-webshell"
fi

# Verify
echo "Verifying installation"
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    VERSION_INFO=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown version")
    echo "WebShell installed successfully"
    echo "Version: ${VERSION_INFO}"
    echo "Location: $(which $BINARY_NAME)"
else
    echo "Installation verification failed"
    echo "Try adding the installation directory to your PATH"
    exit 1
fi

echo ""
echo "WebShell installation completed successfully"
echo ""

# A service install is the bootstrap contract used by application clients. Keeping token, port,
# and path in a root-only EnvironmentFile avoids leaking them through the systemd unit itself.
if [ "$INSTALL_SERVICE" = true ]; then
    if [ -z "${AUTH_TOKEN:-}" ]; then
        echo "AUTH_TOKEN is required when installing the WebShell service"
        exit 1
    fi
    if ! command -v systemctl >/dev/null 2>&1; then
        echo "systemd is required for --service"
        exit 1
    fi
    SERVICE_PORT="${PORT:-8080}"
    SERVICE_PATH="${SECURE_PATH:-/}"
    SERVICE_BINARY="$(command -v "$BINARY_NAME")"
    if [ "$(id -u)" -eq 0 ]; then
        AS_ROOT=""
    else
        AS_ROOT="sudo -n"
    fi
    $AS_ROOT install -d -m 700 /etc/chumen-webshell/tls
    CERT_FILE="/etc/chumen-webshell/tls/server.crt"
    KEY_FILE="/etc/chumen-webshell/tls/server.key"
    if [ ! -s "$CERT_FILE" ] || [ ! -s "$KEY_FILE" ]; then
        command -v openssl >/dev/null 2>&1 || { $AS_ROOT apt-get update -qq && $AS_ROOT apt-get install -y -qq openssl; }
        $AS_ROOT openssl req -x509 -newkey rsa:3072 -sha256 -nodes -days 3650 \
            -subj "/CN=chumen-webshell" -keyout "$KEY_FILE" -out "$CERT_FILE" >/dev/null 2>&1
        $AS_ROOT chmod 600 "$KEY_FILE"
    fi
    TLS_FINGERPRINT=$($AS_ROOT openssl x509 -in "$CERT_FILE" -noout -fingerprint -sha256 | sed 's/.*=//' | tr -d ':')
    $AS_ROOT install -d -m 700 /etc/chumen-webshell
    $AS_ROOT sh -c "umask 077 && cat > /etc/chumen-webshell/chumen-webshell.env" <<EOF
AUTH_TOKEN=${AUTH_TOKEN}
PORT=${SERVICE_PORT}
SECURE_PATH=${SERVICE_PATH}
PUBLIC_PORT=${PUBLIC_PORT:-443}
CERT_FILE=${CERT_FILE}
KEY_FILE=${KEY_FILE}
EOF
    $AS_ROOT tee /etc/systemd/system/${SERVICE_NAME}.service >/dev/null <<EOF
[Unit]
Description=Chumen WebShell management API
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
EnvironmentFile=/etc/chumen-webshell/chumen-webshell.env
ExecStart=${SERVICE_BINARY}
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
    $AS_ROOT systemctl daemon-reload
    $AS_ROOT systemctl enable --now "$SERVICE_NAME"
    $AS_ROOT systemctl is-active --quiet "$SERVICE_NAME"
    echo "WebShell systemd service is active on port ${SERVICE_PORT}"
    echo "CHUMEN_WEBSHELL_TLS_SHA256=${TLS_FINGERPRINT}"
fi

echo "Usage Information:"
echo "Start WebShell: $BINARY_NAME"
echo "Access Web Interface: http://localhost:8080"
echo "Access Web Terminal: http://localhost:8080/terminal"
echo "Documentation: https://github.com/${REPO}"
echo "Custom Port: PORT=3000 $BINARY_NAME"

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "WebShell Install Script"
        echo ""
        echo "Usage: $0 [version]"
        echo ""
        echo "Arguments:"
        echo "  version    Specific version to install (default: latest)"
        echo ""
        echo "Examples:"
        echo "  $0              # Install latest version"
        echo "  $0 v0.1.6       # Install specific version"
        echo "  $0 0.1.6        # Install specific version (without v prefix)"
        exit 0
        ;;
    *)
        ;;
esac
