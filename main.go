package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/adaptive-scale/webshell/internal/agent"
	"github.com/adaptive-scale/webshell/internal/auth"
	"github.com/adaptive-scale/webshell/internal/config"
	"github.com/adaptive-scale/webshell/internal/handler"
	"github.com/adaptive-scale/webshell/internal/terminal"
)

func main() {
	// Parse command line flags
	var (
		port       = flag.String("port", "", "Server port (default: 8080 or PORT env)")
		authToken  = flag.String("token", "", "Authentication token (can also use AUTH_TOKEN env)")
		securePath = flag.String("path", "", "Secure path prefix (default: empty or SECURE_PATH env, e.g., /abc123/)")
		certFile   = flag.String("cert", "", "TLS certificate file (can also use CERT_FILE env)")
		keyFile    = flag.String("key", "", "TLS private key file (can also use KEY_FILE env)")
	)
	flag.Parse()

	// Get port from flag, env, or default
	serverPort := *port
	if serverPort == "" {
		serverPort = config.GetEnv("PORT", "8080")
	}

	// Get auth token from flag or env
	token := *authToken
	if token == "" {
		token = config.GetEnv("AUTH_TOKEN", "")
	}

	if strings.TrimSpace(token) == "" {
		log.Fatal("AUTH_TOKEN is required")
	}
	config.SetAuthToken(token)

	// Get secure path from flag, env, or default (/)
	pathPrefix := *securePath
	if pathPrefix == "" {
		pathPrefix = config.GetEnv("SECURE_PATH", "/")
	}

	// Normalize path prefix: ensure it starts with / and ends with /
	pathPrefix = normalizePath(pathPrefix)

	mux := http.NewServeMux()
	setupRoutes(mux, pathPrefix)

	// Get certificate and key file paths
	certPath := *certFile
	if certPath == "" {
		certPath = config.GetEnv("CERT_FILE", "")
	}
	keyPath := *keyFile
	if keyPath == "" {
		keyPath = config.GetEnv("KEY_FILE", "")
	}
	if certPath == "" || keyPath == "" {
		log.Fatal("CERT_FILE and KEY_FILE are required")
	}
	log.Printf("Path prefix: %s", pathPrefix)
	log.Printf("  - Home: %s", pathPrefix)
	log.Printf("  - Health: %shealth", pathPrefix)
	log.Printf("  - Execute: %sexecute", pathPrefix)
	log.Printf("  - Terminal: %sterminal", pathPrefix)
	log.Printf("  - WebSocket: %sws", pathPrefix)
	log.Printf("  - Upload (POST multipart): %supload", pathPrefix)
	log.Printf("  - Upload (PUT raw body): %supload", pathPrefix)
	log.Printf("  - Download: %sdownload", pathPrefix)

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil {
		log.Fatal("HTTPS server failed:", err)
	}
}

// normalizePath ensures the path starts with / and ends with /
func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}

func setupRoutes(mux *http.ServeMux, pathPrefix string) {
	// Public routes
	mux.HandleFunc(pathPrefix+"{$}", handler.Home)
	mux.HandleFunc(pathPrefix+"health", handler.Health)
	mux.HandleFunc(pathPrefix+"terminal", handler.TerminalPage)

	// Protected routes
	// Agent routes are fixed operations for application clients. Keep them separate from the
	// interactive terminal so Chumen never has to expose an arbitrary command execution surface.
	mux.HandleFunc(pathPrefix+"api/v1/status", auth.AuthMiddleware(agent.Status))
	mux.HandleFunc(pathPrefix+"api/v1/control-plane", auth.AuthMiddleware(agent.ControlPlaneConfiguration))
	mux.HandleFunc(pathPrefix+"api/v1/domain/status", auth.AuthMiddleware(agent.DomainStatus))
	mux.HandleFunc(pathPrefix+"api/v1/domain/setup", auth.AuthMiddleware(agent.SetupDomain))
	mux.HandleFunc(pathPrefix+"api/v1/xray/install", auth.AuthMiddleware(agent.InstallXray))
	mux.HandleFunc(pathPrefix+"execute", auth.AuthMiddleware(handler.ExecuteCommand))
	tickets := terminal.NewTickets()
	mux.HandleFunc(pathPrefix+"terminal-ticket", auth.AuthMiddleware(tickets.Issue))
	mux.HandleFunc(pathPrefix+"ws", tickets.WebSocket)
	mux.HandleFunc(pathPrefix+"upload", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.UploadFile(w, r)
		case http.MethodPut:
			handler.UploadFilePut(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc(pathPrefix+"download", auth.AuthMiddleware(handler.DownloadFile))
}
