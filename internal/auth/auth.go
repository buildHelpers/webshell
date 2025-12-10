package auth

import (
	"net/http"
	"strings"

	"github.com/adaptive-scale/webshell/internal/config"
)

// AuthMiddleware is a middleware function that checks for authentication token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no auth token is configured, allow all requests
		if !config.HasAuthToken() {
			next(w, r)
			return
		}

		// Get token from Authorization header or query parameter
		token := getTokenFromRequest(r)

		// Validate token
		if token == "" || token != config.GetAuthToken() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized", "message": "Invalid or missing authentication token"}`))
			return
		}

		// Token is valid, proceed
		next(w, r)
	}
}

// getTokenFromRequest extracts the token from the request
// Supports:
// 1. Authorization header: "Bearer <token>" or "Token <token>"
// 2. Query parameter: ?token=<token>
// 3. Header: X-Auth-Token
func getTokenFromRequest(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Fields(authHeader)
		if len(parts) == 2 && (parts[0] == "Bearer" || parts[0] == "Token") {
			return parts[1]
		}
	}

	// Check X-Auth-Token header
	if token := r.Header.Get("X-Auth-Token"); token != "" {
		return token
	}

	// Check query parameter
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}

