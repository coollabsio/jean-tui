package github

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager handles GitHub operations using gh CLI
type Manager struct{}

// NewManager creates a new GitHub manager
func NewManager() *Manager {
	return &Manager{}
}

// IsGhInstalled checks if gh CLI is installed
func (m *Manager) IsGhInstalled() bool {
	cmd := exec.Command("gh", "--version")
	return cmd.Run() == nil
}

// IsAuthenticated checks if user is authenticated with gh
func (m *Manager) IsAuthenticated() (bool, error) {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If output contains "not logged into", user is not authenticated
		if strings.Contains(string(output), "not logged into") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check auth status: %s", string(output))
	}
	return true, nil
}

// CreateDraftPR creates a draft pull request
func (m *Manager) CreateDraftPR(worktreePath, branch, baseBranch, title string) (string, error) {
	// Check if gh is installed
	if !m.IsGhInstalled() {
		return "", fmt.Errorf("gh CLI is not installed. Install it from https://cli.github.com")
	}

	// Check if authenticated
	authenticated, err := m.IsAuthenticated()
	if err != nil {
		return "", err
	}
	if !authenticated {
		return "", fmt.Errorf("not authenticated with GitHub. Run 'gh auth login' to authenticate")
	}

	// Create draft PR with title
	args := []string{
		"pr", "create",
		"--draft",
		"--base", baseBranch,
		"--head", branch,
		"--title", title,
		"--body", "", // Empty body for now
	}

	cmd := exec.Command("gh", args...)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %s", string(output))
	}

	// Extract PR URL from output
	prURL := strings.TrimSpace(string(output))
	return prURL, nil
}

// GetRepoName gets the repository name from gh CLI
func (m *Manager) GetRepoName(worktreePath string) (string, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repo name: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
