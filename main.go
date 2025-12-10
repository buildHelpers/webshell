package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

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

	// Set auth token if provided
	if token != "" {
		config.SetAuthToken(token)
		log.Printf("Authentication enabled")
	} else {
		log.Printf("Warning: No authentication token set. Server is open to all requests.")
	}

	// Get secure path from flag, env, or default (/)
	pathPrefix := *securePath
	if pathPrefix == "" {
		pathPrefix = config.GetEnv("SECURE_PATH", "/")
	}

	// Normalize path prefix: ensure it starts with / and ends with /
	pathPrefix = normalizePath(pathPrefix)

	setupRoutes(pathPrefix)

	// Get certificate and key file paths
	certPath := *certFile
	if certPath == "" {
		certPath = config.GetEnv("CERT_FILE", "")
	}
	keyPath := *keyFile
	if keyPath == "" {
		keyPath = config.GetEnv("KEY_FILE", "")
	}

	log.Printf("WebShell server starting on port %s", serverPort)
	if certPath != "" && keyPath != "" {
		log.Printf("HTTPS enabled (cert: %s, key: %s)", certPath, keyPath)
	} else {
		log.Printf("HTTP mode (no certificate provided)")
	}
	log.Printf("Path prefix: %s", pathPrefix)
	log.Printf("  - Home: %s", pathPrefix)
	log.Printf("  - Health: %shealth", pathPrefix)
	log.Printf("  - Execute: %sexecute", pathPrefix)
	log.Printf("  - Terminal: %sterminal", pathPrefix)
	log.Printf("  - WebSocket: %sws", pathPrefix)
	log.Printf("  - Upload: %supload", pathPrefix)
	log.Printf("  - Download: %sdownload", pathPrefix)

	// Start server with or without TLS
	if certPath != "" && keyPath != "" {
		// HTTPS mode
		server := &http.Server{
			Addr: ":" + serverPort,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		if err := server.ListenAndServeTLS(certPath, keyPath); err != nil {
			log.Fatal("HTTPS server failed:", err)
		}
	} else {
		// HTTP mode
		if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
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

func setupRoutes(pathPrefix string) {
	// Public routes
	http.HandleFunc(pathPrefix, handler.Home)
	http.HandleFunc(pathPrefix+"health", handler.Health)

	// Protected routes
	http.HandleFunc(pathPrefix+"execute", auth.AuthMiddleware(handler.ExecuteCommand))
	http.HandleFunc(pathPrefix+"terminal", auth.AuthMiddleware(handler.TerminalPage))
	http.HandleFunc(pathPrefix+"ws", auth.AuthMiddleware(terminal.WebSocket))
	http.HandleFunc(pathPrefix+"upload", auth.AuthMiddleware(handler.UploadFile))
	http.HandleFunc(pathPrefix+"download", auth.AuthMiddleware(handler.DownloadFile))
}
