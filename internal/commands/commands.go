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

// ExecuteCommand executes a command with configurable timeout and returns the result
func ExecuteCommand(command string, args []string, timeoutSeconds int) CommandResponse {
	start := time.Now()

	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)

	// Execute command
	output, err := cmd.CombinedOutput()

	// Calculate duration
	duration := time.Since(start)

	// Get exit code
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		exitCode = -1
	}

	// Prepare response
	response := CommandResponse{
		Success:   err == nil,
		Output:    string(output),
		ExitCode:  exitCode,
		Duration:  duration.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Command:   command + " " + strings.Join(args, " "),
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}
