package config

import (
	"os"
)

var authToken string

// SetAuthToken sets the authentication token
func SetAuthToken(token string) {
	authToken = token
}

// GetAuthToken returns the authentication token
func GetAuthToken() string {
	return authToken
}

// HasAuthToken checks if authentication token is configured
func HasAuthToken() bool {
	return authToken != ""
}

// getEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getHostname returns the system hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
} 