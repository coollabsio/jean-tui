# PR Tracking Implementation Plan

## Overview
Add PR tracking to gcool that saves PR URLs when created through the app and displays them in the worktree details panel, with automatic cleanup on worktree deletion.

## User Requirements
Based on user preferences:
- ✅ Display PR info in **details panel only** (not in main list)
- ✅ Track **PR URL** (simple storage)
- ✅ **Manual tracking only** (save PRs created via gcool's `p` key)
- ✅ **Auto-cleanup** PR data when worktree is deleted

## Current State Analysis

### Existing Infrastructure
- **Config System**: `~/.config/gcool/config.json` stores per-repository settings
- **PR Creation**: Already functional via `p` key using `gh` CLI
- **PR URL Capture**: `prCreatedMsg{prURL: prURL}` already contains URL but only displays temporarily
- **Details View**: `renderDetails()` in `tui/view.go` shows worktree info (branch, path, commit, status)

### Files Affected
1. `config/config.go` - Storage layer
2. `git/worktree.go` - Data model
3. `tui/model.go` - Business logic
4. `tui/view.go` - UI display
5. `tui/update.go` - Cleanup logic

## Implementation Plan

### Phase 1: Storage Layer (`config/config.go`)

#### Add PR Storage to RepoConfig
```go
type RepoConfig struct {
    BaseBranch         string            `json:"base_branch"`
    LastSelectedBranch string            `json:"last_selected_branch,omitempty"`
    Editor             string            `json:"editor,omitempty"`
    AutoFetchInterval  int               `json:"auto_fetch_interval,omitempty"`
    BranchPRs          map[string]string `json:"branch_prs,omitempty"` // NEW: branch -> PR URL
}
```

#### Add Methods to Manager
```go
// SetPRForBranch saves a PR URL for a specific branch
func (m *Manager) SetPRForBranch(repoPath, branch, prURL string) error

// GetPRForBranch retrieves the PR URL for a branch (returns empty string if not found)
func (m *Manager) GetPRForBranch(repoPath, branch string) string

// RemovePRForBranch deletes PR data for a branch
func (m *Manager) RemovePRForBranch(repoPath, branch string) error
```

**Implementation Notes**:
- Initialize `BranchPRs` map if nil before setting values
- Call `save()` after modifications to persist changes
- Handle missing keys gracefully (return empty string, not error)

### Phase 2: Data Model (`git/worktree.go`)

#### Update Worktree Struct
```go
type Worktree struct {
    Path   string
    Branch string
    Commit string
    IsCurrent bool
    PRURL  string  // NEW: PR URL for this branch
}
```

#### Modify List() Method
**Current signature**: `func (m *Manager) List() ([]Worktree, error)`

**Changes needed**:
1. Add `configMgr *config.Manager` parameter: `func (m *Manager) List(configMgr *config.Manager) ([]Worktree, error)`
2. After populating each worktree, look up PR URL:
   ```go
   wt.PRURL = configMgr.GetPRForBranch(m.repoPath, wt.Branch)
   ```

**Impact**: All callers of `List()` need to pass config manager (primarily in `tui/model.go`)

### Phase 3: PR Creation Persistence (`tui/model.go`)

#### Ensure Config Manager Available
Verify `Model` struct has access to config manager (likely already exists as `m.config`)

#### Modify prCreatedMsg Handler (in `tui/update.go`)
**Current location**: Around line 177

**Changes**:
```go
case prCreatedMsg:
    if msg.err != nil {
        m.err = msg.err
        m.status = "Failed to create PR"
    } else {
        m.status = "PR created: " + msg.prURL

        // NEW: Save PR URL to config
        if wt := m.selectedWorktree(); wt != nil {
            if err := m.config.SetPRForBranch(m.repoPath, wt.Branch, msg.prURL); err != nil {
                m.status = "PR created but failed to save: " + err.Error()
            }
        }
    }
    return m, nil
```

**Alternative**: Save in `createPR()` command function in `tui/model.go` (line 301):
```go
func (m Model) createPR(worktree git.Worktree) tea.Cmd {
    return func() tea.Msg {
        // ... existing PR creation logic ...

        prURL, err := m.github.CreateDraftPR(...)
        if err != nil {
            return prCreatedMsg{err: err}
        }

        // Save to config immediately after creation
        if saveErr := m.config.SetPRForBranch(m.repoPath, worktree.Branch, prURL); saveErr != nil {
            // Log error but don't fail the PR creation
            return prCreatedMsg{prURL: prURL, err: fmt.Errorf("PR created but save failed: %w", saveErr)}
        }

        return prCreatedMsg{prURL: prURL}
    }
}
```

**Recommendation**: Save in the Update handler (first approach) for better separation of concerns.

### Phase 4: UI Display (`tui/view.go`)

#### Modify renderDetails() Method
**Current location**: Lines 108-159

**Add PR row** after Status field (around line 143):
```go
details := lipgloss.JoinVertical(lipgloss.Left,
    titleStyle.Render("Selected Worktree"),
    "",
    fmt.Sprintf("%s %s", labelStyle.Render("Branch:"), wt.Branch),
    fmt.Sprintf("%s %s", labelStyle.Render("Path:"), wt.Path),
    fmt.Sprintf("%s %s", labelStyle.Render("Commit:"), wt.Commit),
    fmt.Sprintf("%s %s", labelStyle.Render("Status:"), statusText),

    // NEW: Add PR URL if available
    func() string {
        if wt.PRURL != "" {
            return fmt.Sprintf("%s %s", labelStyle.Render("PR:"), wt.PRURL)
        }
        return ""
    }(),
)
```

**Styling Options**:
- Use different color for PR URL (e.g., blue for links)
- Add icon indicator: `"🔗 " + wt.PRURL`
- Make clickable if terminal supports hyperlinks: `\033]8;;` + URL + `\033\\`

**Example with styling**:
```go
prLinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue

if wt.PRURL != "" {
    prRow = fmt.Sprintf("%s %s",
        labelStyle.Render("PR:"),
        prLinkStyle.Render(wt.PRURL))
}
```

### Phase 5: Auto-Cleanup (`tui/update.go`)

#### Modify Worktree Deletion Handler
**Current location**: Around line 177, `worktreeDeletedMsg` case

**Add cleanup logic**:
```go
case worktreeDeletedMsg:
    if msg.err != nil {
        m.err = msg.err
        m.status = "Failed to delete worktree"
    } else {
        m.status = "Worktree deleted"

        // NEW: Clean up PR data for deleted branch
        if msg.branch != "" {
            if err := m.config.RemovePRForBranch(m.repoPath, msg.branch); err != nil {
                // Log but don't fail - deletion already succeeded
                m.status += " (warning: failed to clean up PR data)"
            }
        }

        m.modal = noModal
    }
    return m, loadWorktrees(m.git)
```

**Required**: Modify `worktreeDeletedMsg` to include branch name:
```go
type worktreeDeletedMsg struct {
    err    error
    branch string  // NEW: branch name for cleanup
}
```

Update `deleteWorktree()` command to return branch name in message.

## Implementation Order

1. **Config Storage** (`config/config.go`)
   - Add `BranchPRs` field
   - Implement `SetPRForBranch()`, `GetPRForBranch()`, `RemovePRForBranch()`

2. **Data Model** (`git/worktree.go`)
   - Add `PRURL` field to `Worktree` struct
   - Update `List()` signature and implementation

3. **Update Callers** (`tui/model.go`)
   - Pass config manager to `git.List()` calls

4. **PR Persistence** (`tui/update.go`)
   - Save PR URL in `prCreatedMsg` handler

5. **UI Display** (`tui/view.go`)
   - Add PR row to `renderDetails()`

6. **Cleanup Logic** (`tui/update.go`)
   - Remove PR data in `worktreeDeletedMsg` handler
   - Update message type and deletion command

## Testing Checklist

### Manual Testing
- [ ] Create PR via `p` key
- [ ] Verify PR URL saved to `~/.config/gcool/config.json`
- [ ] View details panel, confirm PR URL displays
- [ ] Restart gcool, verify PR persists across sessions
- [ ] Delete worktree with PR
- [ ] Verify PR removed from config file
- [ ] Create worktree without PR, verify no errors in details view
- [ ] Test with multiple worktrees/PRs

### Edge Cases
- [ ] Handle missing config file gracefully
- [ ] Handle corrupt config JSON
- [ ] PR creation fails but app continues working
- [ ] Delete worktree when config is read-only
- [ ] Very long PR URLs display correctly
- [ ] Special characters in branch names

### Config File Validation
After implementation, config should look like:
```json
{
  "repositories": {
    "/Users/heyandras/devel/gcool": {
      "base_branch": "main",
      "editor": "code",
      "branch_prs": {
        "feature/add-auth": "https://github.com/coollabsio/gcool/pull/123",
        "fix/bug-123": "https://github.com/coollabsio/gcool/pull/124"
      }
    }
  }
}
```

## Future Enhancements (Out of Scope)

These are NOT part of the current implementation but could be added later:

1. **PR Status Tracking**
   - Query `gh pr view` to get PR state (draft, open, merged, closed)
   - Show status indicator in details (🟡 draft, 🟢 open, 🟣 merged, 🔴 closed)
   - Periodically refresh status

2. **PR Number Storage**
   - Extract and store PR number separately from URL
   - Easier reference: "PR #123" instead of full URL

3. **Open PR in Browser**
   - Add keybinding (e.g., `v` for "view PR")
   - Use `open` (macOS) or `xdg-open` (Linux) to launch browser

4. **Auto-Detection**
   - On load, query `gh pr list --head <branch>` for existing PRs
   - Populate PR data for branches with existing PRs

5. **PR Creation Timestamp**
   - Track when PR was created
   - Show "age" of PR in details

6. **Multi-PR Support**
   - Handle multiple PRs for same branch
   - Show list of related PRs

## Code References

- Config manager: `config/config.go:14-33`
- Worktree struct: `git/worktree.go:14-20`
- PR creation: `tui/model.go:301-354`
- PR created handler: `tui/update.go:177`
- Details view: `tui/view.go:108-159`
- Worktree deletion: `tui/update.go:177` (same handler area)

## Dependencies

No new external dependencies required. Uses existing:
- Standard library (`encoding/json`, `os`, `fmt`)
- Existing `config.Manager`
- Existing `git.Manager`
- Existing `github.Manager`
