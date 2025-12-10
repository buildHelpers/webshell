package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// handleExecuteCommand executes commands via HTTP POST
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

	// Check if JSON response is requested
	acceptHeader := r.Header.Get("Accept")
	wantJSON := acceptHeader == "application/json"

	// Execute command directly without whitelist restriction
	// Support both single commands and full scripts
	// If the command line contains newlines, treat it as a script
	if strings.Contains(commandLine, "\n") {
		// Execute as bash script
		response := commands.ExecuteCommand("bash", []string{"-c", commandLine})
		// Return response based on Accept header
		if wantJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			// Return raw output
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			if response.Success {
				w.Write([]byte(response.Output))
			} else {
				w.Write([]byte("Error: " + response.Error + "\n" + response.Output))
			}
		}
		return
	}

	// Split command into command and arguments for simple commands
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		http.Error(w, "Invalid command", http.StatusBadRequest)
		return
	}

	command := parts[0]
	args := parts[1:]

	// Execute command (no whitelist restriction)
	response := commands.ExecuteCommand(command, args)

	// Return response based on Accept header
	if wantJSON {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// Return raw output
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if response.Success {
			w.Write([]byte(response.Output))
		} else {
			w.Write([]byte("Error: " + response.Error))
		}
	}
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

// UploadFile handles file upload requests
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

	// Get overwrite option (default: skip)
	overwrite := r.FormValue("overwrite") == "true"

	// Check if file exists
	fileExisted := false
	if _, err := os.Stat(targetPath); err == nil {
		fileExisted = true
		if !overwrite {
			// File exists and overwrite is false, skip
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

	// Create or overwrite the file
	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to create file %s: %v", targetPath, err)
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		log.Printf("Failed to write file %s: %v", targetPath, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     "File uploaded successfully",
		"path":        targetPath,
		"filename":    header.Filename,
		"size":        header.Size,
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
