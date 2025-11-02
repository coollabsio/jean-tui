package session

import (
	"os/exec"
	"strings"
	"time"
)

// ClaudeStatus represents the status of Claude in a tmux session
type ClaudeStatus int

const (
	StatusUnknown ClaudeStatus = iota
	StatusBusy
	StatusReady
)

func (s ClaudeStatus) String() string {
	switch s {
	case StatusBusy:
		return "Busy"
	case StatusReady:
		return "Ready"
	default:
		return "Unknown"
	}
}

// StatusDetector detects the status of Claude in a tmux session
type StatusDetector struct {
	sessionName      string
	lastOutput       string
	lastOutputTime   time.Time
	stableThreshold  time.Duration
}

// NewStatusDetector creates a new status detector for a tmux session
func NewStatusDetector(sessionName string) *StatusDetector {
	return &StatusDetector{
		sessionName:     sessionName,
		stableThreshold: 500 * time.Millisecond,
	}
}

// GetStatus returns the current status of Claude in the session
// It captures the last few lines of the tmux pane and detects Claude's state
func (d *StatusDetector) GetStatus() ClaudeStatus {
	output := d.capturePaneContent()
	if output == "" {
		return StatusUnknown
	}

	// Check for busy indicators
	if d.isBusy(output) {
		return StatusBusy
	}

	// Check if output is stable (indicates ready/waiting state)
	if d.isOutputStable(output) {
		return StatusReady
	}

	return StatusUnknown
}

// capturePaneContent captures the last N lines from the tmux pane
func (d *StatusDetector) capturePaneContent() string {
	cmd := exec.Command("tmux", "capture-pane", "-pS", "-5", "-t", d.sessionName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// isBusy checks for Claude activity indicators in the output
func (d *StatusDetector) isBusy(output string) bool {
	busyIndicators := []string{
		"Thinking",
		"Building context",
		"Executing tool",
		"Running",
		"Processing",
	}

	lowerOutput := strings.ToLower(output)
	for _, indicator := range busyIndicators {
		if strings.Contains(lowerOutput, strings.ToLower(indicator)) {
			return true
		}
	}

	// Check if output has changed (activity detected)
	if d.lastOutput != "" && d.lastOutput != output {
		// Output changed - update tracking but don't immediately mark as busy
		// The busy indicators above already catch active processing
		d.lastOutput = output
		d.lastOutputTime = time.Now()
	} else if d.lastOutput == "" {
		// First time seeing output - initialize tracking
		d.lastOutput = output
		d.lastOutputTime = time.Now()
	}

	return false
}

// isOutputStable checks if the output is stable, indicating ready state
func (d *StatusDetector) isOutputStable(output string) bool {
	// Check if output ends with a prompt character
	trimmed := strings.TrimSpace(output)
	if !strings.HasSuffix(trimmed, ">") && !strings.HasSuffix(trimmed, "$") {
		return false
	}

	// Check if output has been stable for the threshold duration
	if d.lastOutput == output && time.Since(d.lastOutputTime) >= d.stableThreshold {
		return true
	}

	return false
}

// Reset clears the cached output state
func (d *StatusDetector) Reset() {
	d.lastOutput = ""
	d.lastOutputTime = time.Now()
}
