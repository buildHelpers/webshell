package agent

// This file owns the small, persisted configuration contract exposed to application clients.
// It deliberately records only WebShell's own public-address metadata; DNS, certificate
// issuance, and proxy-core configuration remain separate future operations.

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const serviceEnvironmentPath = "/etc/chumen-webshell/chumen-webshell.env"

type controlPlaneConfiguration struct {
	Domain          string `json:"domain"`
	Port            int    `json:"port"`
	SecurePath      string `json:"secure_path"`
	TLSEnabled      bool   `json:"tls_enabled"`
	CertificatePath string `json:"certificate_path,omitempty"`
}

type controlPlaneUpdate struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
}

// ControlPlaneConfiguration reads or updates only the public address metadata owned by the
// service. It never accepts an executable path, arbitrary environment value, or shell command.
func ControlPlaneConfiguration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, readControlPlaneConfiguration(serviceEnvironmentPath))
	case http.MethodPut:
		var update controlPlaneUpdate
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&update); err != nil {
			http.Error(w, "Invalid configuration payload", http.StatusBadRequest)
			return
		}
		if err := validateDomainUpdate(update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := updateServiceEnvironment(serviceEnvironmentPath, update); err != nil {
			http.Error(w, "Could not persist WebShell configuration", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, readControlPlaneConfiguration(serviceEnvironmentPath))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func readControlPlaneConfiguration(path string) controlPlaneConfiguration {
	values := readEnvironment(path)
	port, _ := strconv.Atoi(values["PUBLIC_PORT"])
	if port == 0 {
		port = 443
	}
	certificatePath := values["CERT_FILE"]
	return controlPlaneConfiguration{
		Domain:          values["PUBLIC_DOMAIN"],
		Port:            port,
		SecurePath:      values["SECURE_PATH"],
		TLSEnabled:      certificatePath != "" && values["KEY_FILE"] != "",
		CertificatePath: certificatePath,
	}
}

func validateDomainUpdate(update controlPlaneUpdate) error {
	if update.Port < 1 || update.Port > 65535 {
		return fmt.Errorf("Port must be between 1 and 65535")
	}
	domain := strings.TrimSpace(update.Domain)
	if domain == "" {
		return nil // An empty domain clears only the stored public-address metadata.
	}
	if strings.ContainsAny(domain, "/:@\t\r\n ") || len(domain) > 253 || net.ParseIP(domain) != nil {
		return fmt.Errorf("Domain must be a DNS hostname, not a URL or IP address")
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("Domain is not a valid DNS hostname")
		}
	}
	return nil
}

func updateServiceEnvironment(path string, update controlPlaneUpdate) error {
	values := readEnvironment(path)
	values["PUBLIC_DOMAIN"] = strings.TrimSpace(update.Domain)
	values["PUBLIC_PORT"] = strconv.Itoa(update.Port)
	return writeEnvironment(path, values)
}

func readEnvironment(path string) map[string]string {
	values := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return values
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, found := strings.Cut(line, "=")
		if found && key != "" {
			values[key] = value
		}
	}
	return values
}

func writeEnvironment(path string, values map[string]string) error {
	// Keep the security-sensitive token out of HTTP responses, but preserve it on disk.
	keys := []string{"AUTH_TOKEN", "PORT", "SECURE_PATH", "CERT_FILE", "KEY_FILE", "PUBLIC_DOMAIN", "PUBLIC_PORT"}
	var lines []string
	for _, key := range keys {
		if value, ok := values[key]; ok && value != "" {
			lines = append(lines, key+"="+value)
		}
	}
	data := []byte(strings.Join(lines, "\n") + "\n")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".chumen-webshell.env-")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err = temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Chmod(0600); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
