// Package agent exposes a deliberately narrow machine-management API. It is intended for
// authenticated clients such as Chumen; unlike /execute it never accepts caller-provided shell.
package agent

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/adaptive-scale/webshell/internal/commands"
)

const xrayInstallScript = `set -eu
if command -v xray >/dev/null 2>&1; then
  xray version
  exit 0
fi
if ! command -v apt-get >/dev/null 2>&1; then
  echo "Only Debian/Ubuntu hosts are supported by this installer" >&2
  exit 1
fi
apt-get update -qq
apt-get install -y -qq curl
curl -fsSL https://github.com/XTLS/Xray-install/raw/main/install-release.sh -o /tmp/webshell-xray-install.sh
bash /tmp/webshell-xray-install.sh
mkdir -p /usr/local/etc/xray
`

// Status returns service metadata without exposing command execution or host secrets.
func Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"service":   "chumen-webshell",
		"hostname":  hostname(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// InstallXray performs only the built-in installer; its request body is intentionally ignored.
func InstallXray(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	response := commands.ExecuteCommand("/bin/sh", []string{"-c", xrayInstallScript}, 600)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}
