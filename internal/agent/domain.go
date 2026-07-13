package agent

// Domain operations are fixed business actions. Clients send declarative DNS/Caddy parameters;
// WebShell keeps the privileged DNS and reverse-proxy work on the server, not in the Swift app.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type domainSetupRequest struct {
	CloudflareToken string `json:"cloudflare_token"`
	ZoneID          string `json:"zone_id"`
	Hostname        string `json:"hostname"`
	Proxied         bool   `json:"proxied"`
}

func DomainStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	config := readControlPlaneConfiguration(serviceEnvironmentPath)
	writeJSON(w, http.StatusOK, map[string]any{"domain": config.Domain, "port": config.Port, "tls_enabled": config.TLSEnabled, "caddy_installed": commandExists("caddy")})
}

func SetupDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request domainSetupRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || strings.TrimSpace(request.CloudflareToken) == "" || strings.TrimSpace(request.ZoneID) == "" || validateDomainUpdate(controlPlaneUpdate{Domain: request.Hostname, Port: 443}) != nil {
		http.Error(w, "Invalid domain setup payload", http.StatusBadRequest)
		return
	}
	if err := upsertCloudflareARecord(request); err != nil {
		http.Error(w, "Cloudflare DNS setup failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := configureCaddy(request.Hostname); err != nil {
		http.Error(w, "Caddy setup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := updateServiceEnvironment(serviceEnvironmentPath, controlPlaneUpdate{Domain: request.Hostname, Port: 443}); err != nil {
		http.Error(w, "DNS succeeded but configuration could not be saved", http.StatusInternalServerError)
		return
	}
	securePath := readEnvironment(serviceEnvironmentPath)["SECURE_PATH"]
	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "configured",
		"domain":     request.Hostname,
		"public_url": "https://" + request.Hostname + securePath,
	})
}

func commandExists(name string) bool { _, err := exec.LookPath(name); return err == nil }

func upsertCloudflareARecord(request domainSetupRequest) error {
	ip, err := http.Get("https://api.ipify.org")
	if err != nil {
		return err
	}
	defer ip.Body.Close()
	var address string
	if _, err = fmt.Fscan(ip.Body, &address); err != nil {
		return err
	}
	client := &http.Client{}
	endpoint := "https://api.cloudflare.com/client/v4/zones/" + request.ZoneID + "/dns_records?type=A&name=" + request.Hostname
	list, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	list.Header.Set("Authorization", "Bearer "+request.CloudflareToken)
	response, err := client.Do(list)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var existing struct {
		Success bool `json:"success"`
		Result  []struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err = json.NewDecoder(response.Body).Decode(&existing); err != nil {
		return err
	}
	if !existing.Success {
		return fmt.Errorf("%v", existing.Errors)
	}
	payload, _ := json.Marshal(map[string]any{"type": "A", "name": request.Hostname, "content": address, "ttl": 1, "proxied": request.Proxied})
	method, url := http.MethodPost, "https://api.cloudflare.com/client/v4/zones/"+request.ZoneID+"/dns_records"
	if len(existing.Result) > 0 {
		method = http.MethodPut
		url += "/" + existing.Result[0].ID
	}
	write, _ := http.NewRequest(method, url, bytes.NewReader(payload))
	write.Header.Set("Authorization", "Bearer "+request.CloudflareToken)
	write.Header.Set("Content-Type", "application/json")
	updated, err := client.Do(write)
	if err != nil {
		return err
	}
	defer updated.Body.Close()
	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	_ = json.NewDecoder(updated.Body).Decode(&result)
	if !result.Success {
		return fmt.Errorf("%v", result.Errors)
	}
	return nil
}

func configureCaddy(hostname string) error {
	if !commandExists("caddy") {
		if output, err := exec.Command("bash", "-lc", "apt-get update -qq && apt-get install -y -qq caddy").CombinedOutput(); err != nil {
			return fmt.Errorf("%s", strings.TrimSpace(string(output)))
		}
	}
	port, _ := strconv.Atoi(readEnvironment(serviceEnvironmentPath)["PORT"])
	if port < 1025 {
		return fmt.Errorf("chumen-webshell random HTTPS port is unavailable")
	}
	configuration := fmt.Sprintf("%s {\n    reverse_proxy https://127.0.0.1:%d {\n        transport http { tls_insecure_skip_verify }\n    }\n}\n", hostname, port)
	if err := os.WriteFile("/etc/caddy/Caddyfile", []byte(configuration), 0644); err != nil {
		return err
	}
	if output, err := exec.Command("systemctl", "enable", "--now", "caddy").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	if output, err := exec.Command("systemctl", "reload", "caddy").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}
