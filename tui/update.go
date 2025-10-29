package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/gcool/git"
)

// Update handles all state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		// Modal is open - handle modal input
		if m.modal != noModal {
			return m.handleModalInput(msg)
		}

		// Normal mode - handle main UI input
		return m.handleMainInput(msg)

	case worktreesLoadedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to load worktrees", 4*time.Second)
			return m, cmd
		} else {
			m.worktrees = msg.worktrees

			// If we just created a worktree, select it
			if m.lastCreatedBranch != "" {
				for i, wt := range m.worktrees {
					if wt.Branch == m.lastCreatedBranch {
						m.selectedIndex = i
						// Clear the flag
						m.lastCreatedBranch = ""
						break
					}
				}
			} else {
				// Otherwise, restore last selected branch if available
				if m.configManager != nil {
					if lastBranch := m.configManager.GetLastSelectedBranch(m.repoPath); lastBranch != "" {
						// Find the worktree with this branch
						for i, wt := range m.worktrees {
							if wt.Branch == lastBranch {
								m.selectedIndex = i
								break
							}
						}
					}
				}
			}
		}
		return m, nil

	case branchesLoadedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to load branches", 4*time.Second)
			return m, cmd
		} else {
			m.branches = msg.branches
		}
		return m, nil

	case worktreeCreatedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to create worktree", 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Worktree created successfully", 3*time.Second)
			m.modal = noModal

			// Store the newly created branch name for selection after reload
			m.lastCreatedBranch = msg.branch

			// Reload worktrees and select the newly created one
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case worktreeDeletedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to delete worktree", 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Worktree deleted successfully", 3*time.Second)
			m.modal = noModal
			if m.selectedIndex >= len(m.worktrees)-1 {
				m.selectedIndex = len(m.worktrees) - 2
				if m.selectedIndex < 0 {
					m.selectedIndex = 0
				}
			}
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case branchRenamedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to rename branch", 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Branch renamed successfully", 3*time.Second)
			// Rename tmux sessions to match the new branch name
			return m, tea.Batch(
				cmd,
				m.renameSessionsForBranch(msg.oldBranch, msg.newBranch),
				m.loadWorktrees,
			)
		}

	case branchCheckedOutMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to checkout branch", 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Branch checked out successfully", 3*time.Second)
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case baseBranchLoadedMsg:
		m.baseBranch = msg.branch
		return m, nil

	case notificationHideMsg:
		// Only handle if this is the current notification
		if m.notification != nil && m.notification.ID == msg.id {
			m.notificationVisible = false
			return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
				return notificationClearedMsg{id: msg.id}
			})
		}
		return m, nil

	case notificationClearedMsg:
		// Only clear if this is the current notification
		if m.notification != nil && m.notification.ID == msg.id {
			m.notification = nil
		}
		return m, nil

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		return m, nil

	case editorOpenedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to open editor: " + msg.err.Error(), 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Opened in editor", 3*time.Second)
			return m, cmd
		}

	case prCreatedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to create PR: " + msg.err.Error(), 4*time.Second)
			return m, cmd
		} else {
			cmd = m.showSuccessNotification("Draft PR created: " + msg.prURL, 5*time.Second)
			return m, cmd
		}

	case commitCreatedMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to create commit: " + msg.err.Error(), 4*time.Second)
			return m, cmd
		} else {
			// Show success message with commit hash
			if msg.commitHash != "" {
				hashDisplay := msg.commitHash
				if len(msg.commitHash) > 8 {
					hashDisplay = msg.commitHash[:8]
				}
				cmd = m.showSuccessNotification("Commit created: " + hashDisplay, 3*time.Second)
			} else {
				cmd = m.showSuccessNotification("Commit created successfully", 3*time.Second)
			}
			// Refresh worktree list to show clean state
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case branchPulledMsg:
		if msg.err != nil {
			if msg.hadConflict {
				// Show error with abort option
				cmd = m.showWarningNotification("Merge conflict! Run 'git merge --abort' in the worktree to abort.")
				return m, cmd
			} else {
				cmd = m.showErrorNotification("Failed to pull from base branch: " + msg.err.Error(), 5*time.Second)
				return m, cmd
			}
		} else {
			cmd = m.showSuccessNotification("Successfully pulled changes from base branch", 3*time.Second)
			// Refresh worktree list after successful pull
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case refreshWithPullMsg:
		if msg.err != nil {
			cmd = m.showErrorNotification("Failed to refresh: " + msg.err.Error(), 5*time.Second)
			return m, cmd
		} else {
			// Build detailed status message based on what was pulled
			cmd = m.showSuccessNotification(buildRefreshStatusMessage(msg), 3*time.Second)
			// Reload worktree list to show updated status
			return m, tea.Batch(
				cmd,
				m.loadWorktrees,
			)
		}

	case activityTickMsg:
		// Check if enough time has passed since last activity check
		if time.Since(m.lastActivityCheck) >= m.activityCheckInterval {
			m.lastActivityCheck = time.Now()
			cmd = m.checkSessionActivity()
			return m, cmd
		}
		return m, m.scheduleActivityCheck()

	case activityCheckedMsg:
		if msg.err == nil {
			// Update sessions with activity information
			m.sessions = msg.sessions
		}
		// Continue scheduling activity checks
		cmd = m.scheduleActivityCheck()
		return m, cmd
	}

	return m, cmd
}

func (m Model) handleMainInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
			// Save the last selected branch
			if wt := m.selectedWorktree(); wt != nil && m.configManager != nil {
				_ = m.configManager.SetLastSelectedBranch(m.repoPath, wt.Branch)
			}
		}

	case "down", "j":
		if m.selectedIndex < len(m.worktrees)-1 {
			m.selectedIndex++
			// Save the last selected branch
			if wt := m.selectedWorktree(); wt != nil && m.configManager != nil {
				_ = m.configManager.SetLastSelectedBranch(m.repoPath, wt.Branch)
			}
		}

	case "r":
		cmd = m.showInfoNotification("Pulling latest commits and refreshing...")
		return m, tea.Batch(cmd, m.refreshWithPull())

	case "n":
		// Instantly create worktree with random branch name from base branch
		randomName, err := m.gitManager.GenerateRandomName()
		if err != nil {
			cmd = m.showWarningNotification("Failed to generate random name")
			return m, cmd
		}

		// Generate random path
		path, err := m.gitManager.GetDefaultPath(randomName)
		if err != nil {
			cmd = m.showWarningNotification("Failed to generate workspace path")
			return m, cmd
		}

		cmd = m.showInfoNotification("Creating worktree with branch: " + randomName)
		return m, tea.Batch(cmd, m.createWorktree(path, randomName, true))

	case "b":
		// Open change base branch modal (b for base branch)
		m.modal = changeBaseBranchModal
		m.modalFocused = 0
		m.branchIndex = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.filteredBranches = nil
		// Try to find current base branch in the list
		return m, m.loadBranches

	case "a":
		// Open create from existing branch modal
		m.modal = branchSelectModal
		m.modalFocused = 0
		m.createNewBranch = false
		m.branchIndex = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.filteredBranches = nil
		return m, m.loadBranches

	case "d":
		// Open delete modal
		if wt := m.selectedWorktree(); wt != nil && !wt.IsCurrent {
			// Check for uncommitted changes
			hasUncommitted, err := m.gitManager.HasUncommittedChanges(wt.Path)
			if err != nil {
				cmd = m.showWarningNotification("Failed to check for uncommitted changes")
				return m, cmd
			}
			m.deleteHasUncommitted = hasUncommitted
			m.deleteConfirmForce = false
			m.modal = deleteModal
			m.modalFocused = 0
			return m, nil
		} else if wt != nil && wt.IsCurrent {
			return m, m.showWarningNotification("Cannot delete current worktree")
		}

	case "enter":
		// Switch to selected worktree with Claude
		if wt := m.selectedWorktree(); wt != nil {
			// Save the last selected branch before switching
			if m.configManager != nil {
				_ = m.configManager.SetLastSelectedBranch(m.repoPath, wt.Branch)
			}
			m.switchInfo = SwitchInfo{
				Path:         wt.Path,
				Branch:       wt.Branch,
				AutoClaude:   m.autoClaude,
				TerminalOnly: false, // Explicitly use Claude session, not terminal-only
			}
			return m, tea.Quit
		}

	case "R":
		// Rename current branch (Shift+R)
		if wt := m.selectedWorktree(); wt != nil {
			// Check if this is a workspace worktree (in .workspaces directory)
			if !strings.Contains(wt.Path, ".workspaces") {
				return m, m.showWarningNotification("Cannot rename main branch. Only workspace branches can be renamed.")
			}
			m.modal = renameModal
			m.modalFocused = 0
			m.nameInput.SetValue(wt.Branch)
			m.nameInput.Focus()
			m.nameInput.CursorEnd()
			return m, nil
		}

	case "B":
		// Checkout/switch branch in main repository (Shift+B for checkout)
		m.modal = checkoutBranchModal
		m.modalFocused = 0
		m.branchIndex = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.filteredBranches = nil
		return m, m.loadBranches

	case "t":
		// Open terminal in a separate tmux session (not the Claude session)
		if wt := m.selectedWorktree(); wt != nil {
			// Save the last selected branch before switching
			if m.configManager != nil {
				_ = m.configManager.SetLastSelectedBranch(m.repoPath, wt.Branch)
			}
			m.switchInfo = SwitchInfo{
				Path:         wt.Path,
				Branch:       wt.Branch,
				AutoClaude:   false,        // Never auto-start Claude for terminal
				TerminalOnly: true,         // Signal this is a terminal session
			}
			return m, tea.Quit
		}

	case "o":
		// Open worktree in default IDE
		if wt := m.selectedWorktree(); wt != nil {
			cmd = m.showInfoNotification("Opening in editor...")
			return m, tea.Batch(cmd, m.openInEditor(wt.Path))
		}

	case "e":
		// Open editor selection modal
		m.modal = editorSelectModal
		m.modalFocused = 0
		m.editorIndex = 0

		// Find current editor in the list
		if m.configManager != nil {
			currentEditor := m.configManager.GetEditor(m.repoPath)
			for i, editor := range m.editors {
				if editor == currentEditor {
					m.editorIndex = i
					break
				}
			}
		}
		return m, nil

	case "s":
		// Open settings modal
		m.modal = settingsModal
		m.modalFocused = 0
		m.settingsIndex = 0
		return m, nil

	case "S":
		// Open session list modal (Shift+S)
		m.modal = sessionListModal
		m.modalFocused = 0
		m.sessionIndex = 0
		return m, m.loadSessions

	case "p":
		// Pull changes from base branch
		if wt := m.selectedWorktree(); wt != nil {
			// Check if base branch is set
			if m.baseBranch == "" {
				return m, m.showWarningNotification("Base branch not set. Press 'b' to set base branch")
			}

			// Only allow pull if worktree is behind
			if !wt.IsOutdated || wt.BehindCount == 0 {
				return m, m.showInfoNotification("Worktree is already up-to-date with base branch")
			}

			// Don't allow pull on main worktree
			if !strings.Contains(wt.Path, ".workspaces") {
				return m, m.showWarningNotification("Cannot pull on main worktree. Use 'git pull' manually.")
			}

			cmd = m.showInfoNotification("Pulling changes from base branch...")
			return m, tea.Batch(cmd, m.pullFromBaseBranch(wt.Path, m.baseBranch))
		}

	case "P":
		// Create draft PR (push + open PR) (Shift+P)
		if wt := m.selectedWorktree(); wt != nil {
			cmd = m.showInfoNotification("Creating draft PR...")
			return m, tea.Batch(cmd, m.createPR(wt.Path, wt.Branch))
		}

	case "C":
		// Open commit modal (Shift+C for commit)
		if wt := m.selectedWorktree(); wt != nil {
			// Check if worktree has uncommitted changes
			hasUncommitted, err := m.gitManager.HasUncommittedChanges(wt.Path)
			if err != nil {
				return m, m.showErrorNotification("Failed to check for uncommitted changes: " + err.Error(), 3*time.Second)
			}
			if !hasUncommitted {
				cmd = m.showInfoNotification("Nothing to commit - no uncommitted changes in " + wt.Branch)
				return m, cmd
			}
			m.modal = commitModal
			m.modalFocused = 0
			m.commitSubjectInput.SetValue("")
			m.commitSubjectInput.Focus()
			m.commitBodyInput.SetValue("")
			return m, nil
		}

	case "h":
		// Open help modal
		m.modal = helperModal
		return m, nil
	}

	return m, nil
}

func (m Model) handleModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.modal {
	case createModal:
		return m.handleCreateModalInput(msg)

	case deleteModal:
		return m.handleDeleteModalInput(msg)

	case branchSelectModal:
		return m.handleBranchSelectModalInput(msg)

	case checkoutBranchModal:
		return m.handleCheckoutBranchModalInput(msg)

	case sessionListModal:
		return m.handleSessionListModalInput(msg)

	case renameModal:
		return m.handleRenameModalInput(msg)

	case changeBaseBranchModal:
		return m.handleChangeBaseBranchModalInput(msg)

	case editorSelectModal:
		return m.handleEditorSelectModalInput(msg)

	case settingsModal:
		return m.handleSettingsModalInput(msg)

	case tmuxConfigModal:
		return m.handleTmuxConfigModalInput(msg)

	case commitModal:
		return m.handleCommitModalInput(msg)

	case helperModal:
		return m.handleHelperModalInput(msg)
	}

	return m, cmd
}

// handleSearchBasedModalInput is a shared handler for modals with search/filter functionality
// Used by: branchSelectModal, checkoutBranchModal, changeBaseBranchModal
func (m Model) handleSearchBasedModalInput(msg tea.KeyMsg, config searchModalConfig) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.modal = noModal
		m.searchInput.Blur()
		return m, nil

	case "up", "k":
		if m.modalFocused == 0 {
			// In search input, move focus to list
			m.modalFocused = 1
			m.searchInput.Blur()
		} else if m.modalFocused == 1 && m.branchIndex > 0 {
			// In list, move selection up
			m.branchIndex--
		}
		return m, nil

	case "down", "j":
		if m.modalFocused == 0 {
			// In search input, move focus to list
			m.modalFocused = 1
			m.searchInput.Blur()
		} else if m.modalFocused == 1 {
			// In list, move selection down
			branches := m.filteredBranches
			if len(branches) == 0 {
				branches = m.branches
			}
			if m.branchIndex < len(branches)-1 {
				m.branchIndex++
			}
		}
		return m, nil

	case "tab":
		// Cycle: search -> list -> action button -> cancel button -> search
		m.modalFocused = (m.modalFocused + 1) % 4
		if m.modalFocused == 0 {
			m.searchInput.Focus()
		} else {
			m.searchInput.Blur()
		}
		return m, nil

	case "enter":
		if m.modalFocused == 2 {
			// Action button: Execute the configured action
			branch := m.selectedBranch()
			if branch == "" {
				return m, nil
			}
			m.modal = noModal
			m.searchInput.Blur()
			return config.onConfirm(m, branch)
		} else if m.modalFocused == 3 {
			// Cancel button
			m.modal = noModal
			m.searchInput.Blur()
			return m, nil
		} else if m.modalFocused == 0 || m.modalFocused == 1 {
			// In search input or list: move focus to action button
			m.modalFocused = 2
			m.searchInput.Blur()
			return m, nil
		}
	}

	// Handle search input typing
	// Pass all non-navigation keys to search input when in search or list mode
	if m.modalFocused == 0 || m.modalFocused == 1 {
		// Check if this is a navigation key that's already been handled
		key := msg.String()
		isNavigationKey := key == "up" || key == "k" || key == "down" || key == "j" ||
			key == "tab" || key == "enter" || key == "esc"

		if !isNavigationKey {
			// Pass to search input for typing
			m.searchInput.Focus()
			m.searchInput, cmd = m.searchInput.Update(msg)
			// Filter branches based on search
			m.filteredBranches = m.filterBranches(m.searchInput.Value())
			// Reset branch index when filter changes
			m.branchIndex = 0
			m.modalFocused = 0 // Ensure we're tracking search as focused
		}
	}

	return m, cmd
}

// handleListSelectionModalInput is a shared handler for modals with simple list selection
// Used by: sessionListModal, editorSelectModal
func (m Model) handleListSelectionModalInput(msg tea.KeyMsg, config listSelectionConfig) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.modal = noModal
		return m, nil

	case "up", "k":
		if config.getCurrentIndex() > 0 {
			config.decrementIndex(&m)
		}
		return m, nil

	case "down", "j":
		if config.getCurrentIndex() < config.getItemCount(m)-1 {
			config.incrementIndex(&m)
		}
		return m, nil

	case "enter":
		return config.onConfirm(m)

	default:
		// Allow custom key handling (e.g., "d" for delete in session list)
		if config.onCustomKey != nil {
			return config.onCustomKey(m, msg.String())
		}
	}

	return m, nil
}

// searchModalConfig contains configuration for search-based modals
type searchModalConfig struct {
	onConfirm func(m Model, selectedBranch string) (tea.Model, tea.Cmd)
}

// listSelectionConfig contains configuration for list selection modals
type listSelectionConfig struct {
	getCurrentIndex func() int
	getItemCount    func(m Model) int
	incrementIndex  func(m *Model)
	decrementIndex  func(m *Model)
	onConfirm       func(m Model) (tea.Model, tea.Cmd)
	onCustomKey     func(m Model, key string) (tea.Model, tea.Cmd) // Optional
}

func (m Model) handleCreateModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = noModal
		m.nameInput.Blur()
		return m, nil

	case "tab", "shift+tab":
		// Cycle through: nameInput -> create button -> cancel button (if new branch)
		// or just: create button -> cancel button (if existing branch)
		if m.createNewBranch {
			m.modalFocused = (m.modalFocused + 1) % 3
			if m.modalFocused == 0 {
				m.nameInput.Focus()
			} else {
				m.nameInput.Blur()
			}
		} else {
			// For existing branch, just toggle between buttons
			if m.modalFocused == 0 {
				m.modalFocused = 1
			} else if m.modalFocused == 1 {
				m.modalFocused = 2
			} else {
				m.modalFocused = 1
			}
		}
		return m, nil

	case "enter":
		if m.modalFocused <= 1 {
			// Create button or enter in input
			name := m.nameInput.Value()

			if name == "" {
				m.status = "Branch name is required"
				return m, nil
			}

			// Always generate random path
			path, err := m.gitManager.GetDefaultPath(name)
			if err != nil {
				m.status = "Failed to generate workspace path"
				return m, nil
			}

			return m, m.createWorktree(path, name, m.createNewBranch)
		} else if m.modalFocused == 2 {
			// Cancel button
			m.modal = noModal
			m.nameInput.Blur()
			return m, nil
		}
	}

	// Handle text input for branch name
	var cmd tea.Cmd
	if m.modalFocused == 0 && m.createNewBranch {
		m.nameInput, cmd = m.nameInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleDeleteModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.modal = noModal
		return m, nil

	case "tab", "left", "right", "h", "l":
		// If uncommitted changes, we have 3 buttons (Yes/No/Force), otherwise 2 (Yes/No)
		if m.deleteHasUncommitted {
			m.modalFocused = (m.modalFocused + 1) % 3
		} else {
			m.modalFocused = (m.modalFocused + 1) % 2
		}

	case "enter", "y":
		// If there are uncommitted changes and user hasn't confirmed force
		if m.deleteHasUncommitted && !m.deleteConfirmForce {
			// modalFocused: 0 = Yes (blocked), 1 = No, 2 = Force Delete
			if m.modalFocused == 2 || msg.String() == "f" {
				// User clicked "Force Delete" - set confirmation flag
				m.deleteConfirmForce = true
				return m, m.showWarningNotification("Press 'y' or Enter to confirm force delete")
			} else if m.modalFocused == 1 || msg.String() == "n" {
				// User clicked "No" - cancel
				m.modal = noModal
				return m, nil
			} else if m.modalFocused == 0 {
				// User tried to click "Yes" but it's blocked
				return m, m.showWarningNotification("Cannot delete: uncommitted changes. Use 'Force Delete' to proceed.")
			}
		} else if m.deleteHasUncommitted && m.deleteConfirmForce {
			// User already confirmed, now execute force delete
			if m.modalFocused == 0 || msg.String() == "y" {
				if wt := m.selectedWorktree(); wt != nil {
					m.modal = noModal
					return m, m.deleteWorktree(wt.Path, wt.Branch, true) // force = true
				}
			}
			m.modal = noModal
			return m, nil
		} else {
			// No uncommitted changes, normal delete
			if m.modalFocused == 0 || msg.String() == "y" {
				if wt := m.selectedWorktree(); wt != nil {
					m.modal = noModal
					return m, m.deleteWorktree(wt.Path, wt.Branch, false)
				}
			}
			m.modal = noModal
			return m, nil
		}

	case "f":
		// Shortcut for "Force Delete"
		if m.deleteHasUncommitted && !m.deleteConfirmForce {
			m.deleteConfirmForce = true
			return m, m.showWarningNotification("Press 'y' or Enter to confirm force delete")
		}
	}

	return m, nil
}

func (m Model) handleBranchSelectModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := searchModalConfig{
		onConfirm: func(m Model, branch string) (tea.Model, tea.Cmd) {
			// Generate random path
			path, err := m.gitManager.GetDefaultPath(branch)
			if err != nil {
				cmd := m.showWarningNotification("Failed to generate workspace path")
				return m, cmd
			}
			return m, m.createWorktree(path, branch, false)
		},
	}
	return m.handleSearchBasedModalInput(msg, config)
}

func (m Model) handleCheckoutBranchModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := searchModalConfig{
		onConfirm: func(m Model, branch string) (tea.Model, tea.Cmd) {
			cmd := m.showInfoNotification("Checking out branch: " + branch)
			return m, tea.Batch(cmd, m.checkoutBranch(branch))
		},
	}
	return m.handleSearchBasedModalInput(msg, config)
}

func (m Model) handleSessionListModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := listSelectionConfig{
		getCurrentIndex: func() int { return m.sessionIndex },
		getItemCount:    func(m Model) int { return len(m.sessions) },
		incrementIndex:  func(m *Model) { m.sessionIndex++ },
		decrementIndex:  func(m *Model) { m.sessionIndex-- },
		onConfirm: func(m Model) (tea.Model, tea.Cmd) {
			if m.sessionIndex >= 0 && m.sessionIndex < len(m.sessions) {
				sess := m.sessions[m.sessionIndex]
				// Attach via tmux
				if err := m.sessionManager.Attach(sess.Name); err != nil {
					m.showErrorNotification("Failed to attach to session", 3*time.Second)
					return m, nil
				}
				return m, tea.Quit
			}
			return m, nil
		},
		onCustomKey: func(m Model, key string) (tea.Model, tea.Cmd) {
			if key == "d" && m.sessionIndex >= 0 && m.sessionIndex < len(m.sessions) {
				// Kill selected session
				sess := m.sessions[m.sessionIndex]
				if err := m.sessionManager.Kill(sess.Name); err != nil {
					m.showErrorNotification("Failed to kill session", 3*time.Second)
				} else {
					m.showSuccessNotification("Session killed", 3*time.Second)
					// Reload sessions
					return m, m.loadSessions
				}
			}
			return m, nil
		},
	}
	return m.handleListSelectionModalInput(msg, config)
}

func (m Model) handleRenameModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = noModal
		m.nameInput.Blur()
		return m, nil

	case "tab", "shift+tab":
		// Toggle between input and buttons
		m.modalFocused = (m.modalFocused + 1) % 3
		if m.modalFocused == 0 {
			m.nameInput.Focus()
		} else {
			m.nameInput.Blur()
		}
		return m, nil

	case "enter":
		if m.modalFocused <= 1 {
			// Rename button or enter in input
			newName := m.nameInput.Value()
			if newName == "" {
				cmd := m.showWarningNotification("Branch name cannot be empty")
				return m, cmd
			}

			// Sanitize branch name
			newName = git.SanitizeBranchName(newName)
			if newName == "" {
				cmd := m.showWarningNotification("Branch name cannot be empty after sanitization")
				return m, cmd
			}

			if wt := m.selectedWorktree(); wt != nil {
				if newName == wt.Branch {
					cmd := m.showInfoNotification("Branch name unchanged")
					m.modal = noModal
					m.nameInput.Blur()
					return m, cmd
				}

				cmd := m.showInfoNotification(fmt.Sprintf("Renaming branch to '%s'...", newName))
				m.modal = noModal
				m.nameInput.Blur()
				return m, tea.Batch(cmd, m.renameBranch(wt.Branch, newName))
			}
		} else if m.modalFocused == 2 {
			// Cancel button
			m.modal = noModal
			m.nameInput.Blur()
			return m, nil
		}
	}

	// Handle text input
	var cmd tea.Cmd
	if m.modalFocused == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleChangeBaseBranchModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := searchModalConfig{
		onConfirm: func(m Model, branch string) (tea.Model, tea.Cmd) {
			m.baseBranch = branch
			var cmd tea.Cmd

			// Save to config
			if m.configManager != nil {
				if err := m.configManager.SetBaseBranch(m.repoPath, branch); err != nil {
					cmd = m.showWarningNotification("Base branch set to: " + branch + " (warning: failed to save)")
				} else {
					cmd = m.showSuccessNotification("Base branch set to: " + branch, 3*time.Second)
				}
			} else {
				cmd = m.showInfoNotification("Base branch set to: " + branch)
			}
			return m, cmd
		},
	}
	return m.handleSearchBasedModalInput(msg, config)
}

func (m Model) handleCommitModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = noModal
		m.commitSubjectInput.Blur()
		m.commitBodyInput.Blur()
		return m, nil

	case "tab", "shift+tab":
		// Cycle through: subject input -> body input -> commit button -> cancel button
		m.modalFocused = (m.modalFocused + 1) % 4

		// Update focus state
		if m.modalFocused == 0 {
			m.commitSubjectInput.Focus()
			m.commitBodyInput.Blur()
		} else if m.modalFocused == 1 {
			m.commitSubjectInput.Blur()
			m.commitBodyInput.Focus()
		} else {
			m.commitSubjectInput.Blur()
			m.commitBodyInput.Blur()
		}
		return m, nil

	case "enter":
		if m.modalFocused == 0 {
			// In subject input, move to body
			m.modalFocused = 1
			m.commitSubjectInput.Blur()
			m.commitBodyInput.Focus()
			return m, nil
		} else if m.modalFocused == 1 {
			// In body input, move to commit button
			m.modalFocused = 2
			m.commitBodyInput.Blur()
			return m, nil
		} else if m.modalFocused == 2 {
			// Commit button
			subject := m.commitSubjectInput.Value()
			if subject == "" {
				cmd := m.showWarningNotification("Commit subject cannot be empty")
				return m, cmd
			}

			body := m.commitBodyInput.Value()
			if wt := m.selectedWorktree(); wt != nil {
				cmd := m.showInfoNotification("Creating commit...")
				m.modal = noModal
				m.commitSubjectInput.Blur()
				m.commitBodyInput.Blur()
				return m, tea.Batch(cmd, m.createCommit(wt.Path, subject, body))
			}
		} else {
			// Cancel button
			m.modal = noModal
			m.commitSubjectInput.Blur()
			m.commitBodyInput.Blur()
			return m, nil
		}
	}

	// Handle text input
	var cmd tea.Cmd
	if m.modalFocused == 0 {
		m.commitSubjectInput, cmd = m.commitSubjectInput.Update(msg)
	} else if m.modalFocused == 1 {
		m.commitBodyInput, cmd = m.commitBodyInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleEditorSelectModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	config := listSelectionConfig{
		getCurrentIndex: func() int { return m.editorIndex },
		getItemCount:    func(m Model) int { return len(m.editors) },
		incrementIndex:  func(m *Model) { m.editorIndex++ },
		decrementIndex:  func(m *Model) { m.editorIndex-- },
		onConfirm: func(m Model) (tea.Model, tea.Cmd) {
			if m.editorIndex >= 0 && m.editorIndex < len(m.editors) {
				selectedEditor := m.editors[m.editorIndex]
				if m.configManager != nil {
					if err := m.configManager.SetEditor(m.repoPath, selectedEditor); err != nil {
						m.showErrorNotification("Failed to save editor preference", 3*time.Second)
					} else {
						m.showSuccessNotification("Editor set to: " + selectedEditor, 3*time.Second)
					}
				}
			}
			m.modal = noModal
			return m, nil
		},
	}
	return m.handleListSelectionModalInput(msg, config)
}

func (m Model) handleSettingsModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.modal = noModal
		return m, nil

	case "up", "k":
		if m.settingsIndex > 0 {
			m.settingsIndex--
		}

	case "down", "j":
		if m.settingsIndex < 2 { // Now 3 settings (editor, base branch, tmux config)
			m.settingsIndex++
		}

	case "enter":
		// Open the selected setting's modal
		switch m.settingsIndex {
		case 0:
			// Editor setting - open editor select modal
			m.modal = editorSelectModal
			m.modalFocused = 0
			m.editorIndex = 0

			// Find current editor in the list
			if m.configManager != nil {
				currentEditor := m.configManager.GetEditor(m.repoPath)
				for i, editor := range m.editors {
					if editor == currentEditor {
						m.editorIndex = i
						break
					}
				}
			}
			return m, nil

		case 1:
			// Base branch setting - open change base branch modal
			m.modal = changeBaseBranchModal
			m.modalFocused = 0
			m.branchIndex = 0
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			m.filteredBranches = nil
			return m, m.loadBranches

		case 2:
			// Tmux config setting - open tmux config modal
			m.modal = tmuxConfigModal
			m.modalFocused = 0
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleTmuxConfigModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.modal = settingsModal
		return m, nil

	case "tab", "shift+tab":
		// Check if config exists to determine button count
		hasConfig := false
		if m.sessionManager != nil {
			installed, _ := m.sessionManager.HasGcoolTmuxConfig()
			hasConfig = installed
		}

		// If config exists: Update (0), Remove (1), Cancel (2) = 3 buttons
		// If not exists: Install (0), Cancel (1) = 2 buttons
		if hasConfig {
			m.modalFocused = (m.modalFocused + 1) % 3
		} else {
			m.modalFocused = (m.modalFocused + 1) % 2
		}
		return m, nil

	case "enter":
		if m.sessionManager == nil {
			m.modal = settingsModal
			return m, nil
		}

		hasConfig, err := m.sessionManager.HasGcoolTmuxConfig()
		if err != nil {
			m.showErrorNotification("Error checking tmux config: " + err.Error(), 3*time.Second)
			m.modal = settingsModal
			return m, nil
		}

		if hasConfig {
			// Config exists: Update (0), Remove (1), Cancel (2)
			switch m.modalFocused {
			case 0:
				// Update button - reinstalls config (remove + add)
				if err := m.sessionManager.AddGcoolTmuxConfig(); err != nil {
					m.showErrorNotification("Failed to update tmux config: " + err.Error(), 3*time.Second)
				} else {
					m.showSuccessNotification("gcool tmux config updated! New tmux sessions will use the updated config.", 3*time.Second)
				}
			case 1:
				// Remove button
				if err := m.sessionManager.RemoveGcoolTmuxConfig(); err != nil {
					m.showErrorNotification("Failed to remove tmux config: " + err.Error(), 3*time.Second)
				} else {
					m.showSuccessNotification("gcool tmux config removed. New tmux sessions will use your default config.", 3*time.Second)
				}
			case 2:
				// Cancel button - do nothing
			}
		} else {
			// Config doesn't exist: Install (0), Cancel (1)
			switch m.modalFocused {
			case 0:
				// Install button
				if err := m.sessionManager.AddGcoolTmuxConfig(); err != nil {
					m.showErrorNotification("Failed to add tmux config: " + err.Error(), 3*time.Second)
				} else {
					m.showSuccessNotification("gcool tmux config installed! New tmux sessions will use this config.", 3*time.Second)
				}
			case 1:
				// Cancel button - do nothing
			}
		}

		// Return to settings modal
		m.modal = settingsModal
		return m, nil
	}

	return m, nil
}

// buildRefreshStatusMessage constructs a detailed status message based on refresh results
func buildRefreshStatusMessage(msg refreshWithPullMsg) string {
	// If everything was already up to date
	if msg.upToDate && len(msg.updatedBranches) == 0 && !msg.mergedBaseBranch {
		return "Already up to date (0 new commits)"
	}

	// Build a summary with commit counts
	var totalCommits int
	var branchDetails []string

	// Add branch-specific updates
	for branch, commits := range msg.updatedBranches {
		if commits > 0 {
			totalCommits += commits
			branchDetails = append(branchDetails, fmt.Sprintf("%s (+%d)", branch, commits))
		}
	}

	// If we merged base branch into worktree, note that
	if msg.mergedBaseBranch && len(branchDetails) > 0 {
		return fmt.Sprintf("Pulled %d commits: %s", totalCommits, strings.Join(branchDetails, ", "))
	} else if msg.mergedBaseBranch {
		return "Merged base branch into worktree"
	}

	// Summary message
	if totalCommits == 0 {
		return "Refreshed (no new commits)"
	}

	if len(branchDetails) == 1 {
		return fmt.Sprintf("Pulled %d new commits in %s", totalCommits, branchDetails[0])
	}

	return fmt.Sprintf("Pulled %d commits: %s", totalCommits, strings.Join(branchDetails, ", "))
}

func (m Model) handleHelperModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "q":
		// Close helper modal with Esc, h, or q
		m.modal = noModal
		return m, nil
	}

	return m, nil
}
