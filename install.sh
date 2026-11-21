#!/bin/bash

# WebShell Install Script
set -euo pipefail

# Configuration
REPO="buildHelpers/webshell"
VERSION="v1.2.5"
BINARY_NAME="chumen-webshell"
SERVICE_NAME="chumen-webshell"
INSTALL_SERVICE=false

for argument in "$@"; do
    case "$argument" in
        -h|--help)
            echo "Usage: $0 [version] [--service]"
            echo "Default release: $VERSION. Service installation requires AUTH_TOKEN."
            exit 0
            ;;
        --service) INSTALL_SERVICE=true ;;
        *) VERSION="$argument" ;;
    esac
done

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

VERSION="v${VERSION#v}"
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Specify an exact WebShell release, for example v1.2.5." >&2
    exit 1
fi
SERVICE_PORT="${PORT:-8080}"
SERVICE_PATH="${SECURE_PATH:-/}"
PUBLIC_PORT="${PUBLIC_PORT:-443}"
if [ "$INSTALL_SERVICE" = true ]; then
    if [[ ! "${AUTH_TOKEN:-}" =~ [^[:space:]] || "$AUTH_TOKEN" =~ [[:cntrl:]] ]]; then
        echo "A nonempty AUTH_TOKEN without control characters is required." >&2
        exit 1
    fi
    if [[ ! "$SERVICE_PORT" =~ ^[0-9]{1,5}$ || ! "$PUBLIC_PORT" =~ ^[0-9]{1,5}$ ||
          ! "$SERVICE_PATH" =~ ^/[a-zA-Z0-9/_-]*$ ]]; then
        echo "Invalid WebShell service port or path." >&2
        exit 1
    fi
    if ! command -v systemctl >/dev/null 2>&1; then
        echo "systemd is required for --service." >&2
        exit 1
    fi
fi

FILENAME="webshell_$PLATFORM"
TMP_DIR="$(mktemp -d)"
trap 'rm -f "$TMP_DIR/$FILENAME"; rmdir "$TMP_DIR"' EXIT

# GitHub returns the digest for each asset in a published immutable release.
DIGEST="$(curl -fsSL --proto '=https' \
    "https://api.github.com/repos/$REPO/releases/tags/$VERSION" |
    python3 -c '
import json, re, sys
release = json.load(sys.stdin)
tag, name = sys.argv[1:]
if release["tag_name"] != tag or release["immutable"] is not True:
    raise SystemExit("WebShell requires an immutable release")
for asset in release["assets"]:
    if asset["name"] == name and re.fullmatch(r"sha256:[a-f0-9]{64}", asset.get("digest") or ""):
        print(asset["digest"][7:])
        break
else:
    raise SystemExit("Release asset has no GitHub SHA-256 digest")
' "$VERSION" "$FILENAME")"
curl -fsSL --proto '=https' \
    "https://github.com/$REPO/releases/download/$VERSION/$FILENAME" -o "$TMP_DIR/$FILENAME"
ACTUAL="$(openssl dgst -sha256 -r "$TMP_DIR/$FILENAME" | cut -d ' ' -f 1)"
if [ "$ACTUAL" != "$DIGEST" ]; then
    echo "WebShell release asset checksum mismatch." >&2
    exit 1
fi

as_root() {
    if [ "$(id -u)" -eq 0 ]; then
        "$@"
    else
        sudo -n "$@"
    fi
}

SERVICE_BINARY="/usr/local/bin/$BINARY_NAME"
if [ "$INSTALL_SERVICE" = true ] || [ "$(id -u)" -eq 0 ] || sudo -n true 2>/dev/null; then
    as_root install -m 755 "$TMP_DIR/$FILENAME" "$SERVICE_BINARY"
else
    mkdir -p "$HOME/.local/bin"
    SERVICE_BINARY="$HOME/.local/bin/$BINARY_NAME"
    install -m 755 "$TMP_DIR/$FILENAME" "$SERVICE_BINARY"
fi
echo "Installed verified WebShell $VERSION to $SERVICE_BINARY"

# A service install is the bootstrap contract used by application clients. Keeping token, port,
# and path in a root-only EnvironmentFile avoids leaking them through the systemd unit itself.
if [ "$INSTALL_SERVICE" = true ]; then
    as_root install -d -m 700 /etc/chumen-webshell/tls
    CERT_FILE="/etc/chumen-webshell/tls/server.crt"
    KEY_FILE="/etc/chumen-webshell/tls/server.key"
    if ! as_root test -s "$CERT_FILE" || ! as_root test -s "$KEY_FILE"; then
        as_root openssl req -x509 -newkey rsa:3072 -sha256 -nodes -days 3650 \
            -subj "/CN=chumen-webshell" -keyout "$KEY_FILE" -out "$CERT_FILE" >/dev/null 2>&1
        as_root chmod 600 "$KEY_FILE"
    fi
    TLS_FINGERPRINT=$(as_root openssl x509 -in "$CERT_FILE" -noout -fingerprint -sha256 | sed 's/.*=//' | tr -d ':')
    as_root install -d -m 700 /etc/chumen-webshell
    as_root install -m 600 /dev/null /etc/chumen-webshell/chumen-webshell.env
    ESCAPED_TOKEN="$(printf '%s' "$AUTH_TOKEN" | sed 's/[\\"]/\\&/g')"
    as_root tee /etc/chumen-webshell/chumen-webshell.env >/dev/null <<EOF
AUTH_TOKEN="${ESCAPED_TOKEN}"
PORT=${SERVICE_PORT}
SECURE_PATH=${SERVICE_PATH}
PUBLIC_PORT=${PUBLIC_PORT}
CERT_FILE=${CERT_FILE}
KEY_FILE=${KEY_FILE}
EOF
    as_root tee /etc/systemd/system/${SERVICE_NAME}.service >/dev/null <<EOF
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
    as_root systemctl daemon-reload
    as_root systemctl enable "$SERVICE_NAME"
    as_root systemctl restart "$SERVICE_NAME"
    as_root systemctl is-active --quiet "$SERVICE_NAME"
    echo "WebShell systemd service is active on port ${SERVICE_PORT}"
    echo "CHUMEN_WEBSHELL_TLS_SHA256=${TLS_FINGERPRINT}"
fi

echo "Usage Information:"
echo "Start WebShell with AUTH_TOKEN, CERT_FILE and KEY_FILE configured."
echo "Fixed terminal entry: https://<host>:${SERVICE_PORT}${SERVICE_PATH%/}/terminal"
echo "Documentation: https://github.com/${REPO}"
