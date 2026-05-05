package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adaptive-scale/webshell/internal/commands"
	"github.com/adaptive-scale/webshell/internal/config"
	"github.com/adaptive-scale/webshell/internal/templates"
)

// handleHome serves the home page with usage information
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := templates.GetHomeTemplate()
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Failed to get home template: %v", err)
		return
	}

	// No command whitelist - all commands are allowed
	data := struct {
		AllowedCommands []string
	}{
		AllowedCommands: []string{}, // Empty list - all commands are allowed
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// ExecuteCommand executes commands via HTTP POST
// Always uses bash -c for consistent shell behavior
// Always returns JSON response with exit_code
// Supports X-Timeout header for client-specified timeout (default: 300s)
func ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse command from raw body
	commandLine := strings.TrimSpace(string(body))
	if commandLine == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	// Parse timeout from X-Timeout header (seconds), default 300
	timeoutSeconds := 300
	if timeoutStr := r.Header.Get("X-Timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeoutSeconds = t
		}
	}

	// Always execute via bash -c for consistent shell behavior
	// This ensures env vars, redirections, pipes all work correctly
	response := commands.ExecuteCommand("bash", []string{"-c", commandLine}, timeoutSeconds)

	// Always return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleHealth serves the health check endpoint
func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    "running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleTerminalPage serves the web terminal page
func TerminalPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := templates.GetTerminalTemplate()
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Failed to get terminal template: %v", err)
		return
	}

	data := struct {
		Hostname string
	}{
		Hostname: config.GetHostname(),
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// validatePath checks for path traversal attacks
func validatePath(path string) error {
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	if !filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute path required")
	}
	return nil
}

// UploadFile handles file upload requests via multipart/form-data (POST)
func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		log.Printf("Failed to parse multipart form: %v", err)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		log.Printf("No file in request: %v", err)
		return
	}
	defer file.Close()

	// Get target path from form
	targetPath := r.FormValue("path")
	if targetPath == "" {
		http.Error(w, "Target path is required", http.StatusBadRequest)
		return
	}

	// Validate path
	if err := validatePath(targetPath); err != nil {
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	// Get overwrite option (default: skip)
	overwrite := r.FormValue("overwrite") == "true"

	// Check if file exists
	fileExisted := false
	if _, err := os.Stat(targetPath); err == nil {
		fileExisted = true
		if !overwrite {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "skipped",
				"message": "File already exists, skipped",
				"path":    targetPath,
			})
			return
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to create directory %s: %v", dir, err)
		return
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to read uploaded file: %v", err)
		return
	}

	// Write file and explicitly set permissions (bypass umask)
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to write file %s: %v", targetPath, err)
		return
	}
	os.Chmod(targetPath, 0644)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     "File uploaded successfully",
		"path":        targetPath,
		"filename":    header.Filename,
		"size":        len(content),
		"overwritten": fileExisted,
	})
}

// UploadFilePut handles file upload via PUT request
// PUT /upload?path=/tmp/config.json&overwrite=true
// Body contains raw file content (no multipart encoding needed)
func UploadFilePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get target path from query parameter
	targetPath := r.URL.Query().Get("path")
	if targetPath == "" {
		http.Error(w, "Target path is required (use ?path=...)", http.StatusBadRequest)
		return
	}

	// Validate path
	if err := validatePath(targetPath); err != nil {
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	// Get overwrite option (default: true for PUT semantics)
	overwriteStr := r.URL.Query().Get("overwrite")
	overwrite := overwriteStr != "false" // default true

	// Check if file exists
	fileExisted := false
	if _, err := os.Stat(targetPath); err == nil {
		fileExisted = true
		if !overwrite {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "skipped",
				"message": "File already exists, skipped",
				"path":    targetPath,
			})
			return
		}
	}

	// Read request body
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		log.Printf("Failed to read PUT body: %v", err)
		return
	}
	defer r.Body.Close()

	if len(content) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to create directory %s: %v", dir, err)
		return
	}

	// Write file and explicitly set permissions (bypass umask)
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to write file %s: %v", targetPath, err)
		return
	}
	os.Chmod(targetPath, 0644)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     "File uploaded successfully",
		"path":        targetPath,
		"size":        len(content),
		"overwritten": fileExisted,
	})
}

// DownloadFile handles file download requests
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get file path from query parameter
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Validate path
	if err := validatePath(filePath); err != nil {
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to access file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Check if it's a directory
	if info.IsDir() {
		http.Error(w, "Path is a directory, not a file", http.StatusBadRequest)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers for file download
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// Copy file to response
	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("Failed to send file %s: %v", filePath, err)
		return
	}
}
