package commands

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// CommandResponse represents the structure of command execution responses
type CommandResponse struct {
	Success   bool   `json:"success"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
}

// executeCommand executes a command with timeout and returns the result
// Timeout is increased to 300 seconds to support script execution
func ExecuteCommand(command string, args []string) CommandResponse {
	start := time.Now()

	// Create context with timeout (300 seconds for script execution)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)

	// Execute command
	output, err := cmd.CombinedOutput()

	// Calculate duration
	duration := time.Since(start)

	// Prepare response
	response := CommandResponse{
		Success:   err == nil,
		Output:    string(output), // Keep full output including newlines
		ExitCode:  cmd.ProcessState.ExitCode(),
		Duration:  duration.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Command:   command + " " + strings.Join(args, " "),
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}
