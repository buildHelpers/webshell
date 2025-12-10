package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/adaptive-scale/webshell/internal/auth"
	"github.com/adaptive-scale/webshell/internal/config"
	"github.com/adaptive-scale/webshell/internal/handler"
	"github.com/adaptive-scale/webshell/internal/terminal"
)

func main() {
	// Parse command line flags
	var (
		port      = flag.String("port", "", "Server port (default: 8080 or PORT env)")
		authToken = flag.String("token", "", "Authentication token (can also use AUTH_TOKEN env)")
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

	setupRoutes()

	log.Printf("WebShell server starting on port %s", serverPort)

	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func setupRoutes() {
	// Public routes (no authentication required)
	http.HandleFunc("/", handler.Home)
	http.HandleFunc("/health", handler.Health)

	// Protected routes (authentication required if token is set)
	http.HandleFunc("/execute", auth.AuthMiddleware(handler.ExecuteCommand))
	http.HandleFunc("/terminal", auth.AuthMiddleware(handler.TerminalPage))
	http.HandleFunc("/ws", auth.AuthMiddleware(terminal.WebSocket))
}
