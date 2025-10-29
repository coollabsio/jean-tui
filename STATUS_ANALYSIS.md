# Comprehensive Analysis of m.status and m.err Assignments in tui/update.go

## Summary Statistics
- **Total m.status assignments:** 38
- **Total m.err assignments:** 21
- **Message types identified:** 4 (Success, Error, Warning, Info)
- **Auto-clear patterns:** Multiple (tea.Tick with various durations)

---

## PART 1: m.status Assignments with Context

### ✅ SUCCESS MESSAGES (Auto-clear: 2-5 seconds)

| Line | Assignment | Duration | Context | Message Type |
|------|-----------|----------|---------|--------------|
| 35 | `m.status = "Failed to load worktrees"` | 4s | worktreesLoadedMsg error | Error |
| 43 | `m.status = ""` | N/A | worktreesLoadedMsg success, clear | N/A |
| 75 | `m.status = "Failed to load branches"` | 4s | branchesLoadedMsg error | Error |
| 95 | `m.status = "Worktree created successfully"` | 2s | worktreeCreatedMsg success | Success |
| 117 | `m.status = "Failed to delete worktree"` | 4s | worktreeDeletedMsg error | Error |
| 123 | `m.status = "Worktree deleted successfully"` | N/A | worktreeDeletedMsg success, no auto-clear | Success |
| 138 | `m.status = "Failed to rename branch"` | 4s | branchRenamedMsg error | Error |
| 144 | `m.status = "Branch renamed successfully"` | N/A | branchRenamedMsg success, no auto-clear | Success |
| 156 | `m.status = "Failed to checkout branch"` | 4s | branchCheckedOutMsg error | Error |
| 162 | `m.status = "Branch checked out successfully"` | N/A | branchCheckedOutMsg success, no auto-clear | Success |
| 187 | `m.status = "Failed to open editor: " + msg.err.Error()` | 4s | editorOpenedMsg error | Error |
| 193 | `m.status = "Opened in editor"` | 2s | editorOpenedMsg success | Success |
| 204 | `m.status = "Failed to create PR: " + msg.err.Error()` | 4s | prCreatedMsg error | Error |
| 210 | `m.status = "Draft PR created: " + msg.prURL` | 5s | prCreatedMsg success | Success |
| 221 | `m.status = "Failed to create commit: " + msg.err.Error()` | 4s | commitCreatedMsg error | Error |
| 233 | `m.status = "Commit created: " + hashDisplay` | 3s | commitCreatedMsg success (with hash) | Success |
| 235 | `m.status = "Commit created successfully"` | 3s | commitCreatedMsg success (no hash) | Success |
| 254 | `m.status = "Merge conflict! Run 'git merge --abort' in the worktree to abort."` | 5s | branchPulledMsg conflict | Error/Warning |
| 256 | `m.status = "Failed to pull from base branch: " + msg.err.Error()` | 5s | branchPulledMsg error | Error |
| 263 | `m.status = "Successfully pulled changes from base branch"` | 2s | branchPulledMsg success | Success |
| 279 | `m.status = "Failed to refresh: " + msg.err.Error()` | 5s | refreshWithPullMsg error | Error |
| 286 | `m.status = buildRefreshStatusMessage(msg)` | 2s | refreshWithPullMsg success | Success |
| 177 | `m.status = string(msg)` | N/A | statusMsg (generic message passthrough) | Generic |

### ⚠️ TRANSIENT MESSAGES (Set but not explicitly cleared by auto-clear)

| Line | Assignment | Duration | Context | Message Type |
|------|-----------|----------|---------|--------------|
| 345 | `m.status = "Pulling latest commits and refreshing..."` | N/A | User pressed 'r' for refresh | Info |
| 352 | `m.status = "Failed to generate random name"` | N/A | 'n' keybinding - random name generation failed | Error |
| 359 | `m.status = "Failed to generate workspace path"` | N/A | 'n' keybinding - path generation failed | Error |
| 363 | `m.status = "Creating worktree with branch: " + randomName` | N/A | 'n' keybinding - worktree creation started | Info |
| 394 | `m.status = "Failed to check for uncommitted changes"` | N/A | 'd' keybinding - delete check failed | Error |
| 403 | `m.status = "Cannot delete current worktree"` | N/A | 'd' keybinding - safety check | Warning |
| 427 | `m.status = "Cannot rename main branch. Only workspace branches can be renamed."` | N/A | 'R' keybinding - rename protection | Warning |
| 467 | `m.status = "Opening in editor..."` | N/A | 'o' keybinding - editor opening initiated | Info |
| 508 | `m.status = "Base branch not set. Press 'c' to set base branch"` | N/A | 'p' keybinding - pull without base branch | Error/Warning |
| 514 | `m.status = "Worktree is already up-to-date with base branch"` | N/A | 'p' keybinding - no pull needed | Info |
| 520 | `m.status = "Cannot pull on main worktree. Use 'git pull' manually."` | N/A | 'p' keybinding - safety check | Warning |
| 524 | `m.status = "Pulling changes from base branch..."` | N/A | 'p' keybinding - pull initiated | Info |
| 531 | `m.status = "Creating draft PR..."` | N/A | 'P' keybinding - PR creation initiated | Info |
| 542 | `m.status = "Failed to check for uncommitted changes: " + err.Error()` | 3s | 'C' keybinding - commit check failed | Error |
| 549 | `m.status = "Nothing to commit - no uncommitted changes in " + wt.Branch` | 3s | 'C' keybinding - no changes to commit | Info/Warning |
| 794 | `m.status = "Branch name is required"` | N/A | Create modal - empty branch name | Error |
| 801 | `m.status = "Failed to generate workspace path"` | N/A | Create modal - path generation failed | Error |
| 844 | `m.status = "Press 'y' or Enter to confirm force delete"` | N/A | Delete modal - force delete confirmation needed | Warning |
| 852 | `m.status = "Cannot delete: uncommitted changes. Use 'Force Delete' to proceed."` | N/A | Delete modal - safety check | Warning |
| 895 | `m.status = "Failed to generate workspace path"` | N/A | Branch select modal - path generation failed | Error |
| 907 | `m.status = "Checking out branch: " + branch` | N/A | Checkout branch modal - checkout initiated | Info |
| 925 | `m.status = "Failed to attach to session"` | N/A | Session list modal - attach failed | Error |
| 937 | `m.status = "Failed to kill session"` | N/A | Session list modal - kill failed | Error |
| 939 | `m.status = "Session killed"` | N/A | Session list modal - session killed | Success |
| 972 | `m.status = "Branch name cannot be empty"` | N/A | Rename modal - empty name validation | Error |
| 979 | `m.status = "Branch name cannot be empty after sanitization"` | N/A | Rename modal - sanitization validation | Error |
| 985 | `m.status = "Branch name unchanged"` | N/A | Rename modal - name didn't change | Info |
| 991 | `m.status = fmt.Sprintf("Renaming branch to '%s'...", newName)` | N/A | Rename modal - rename initiated | Info |
| 1021 | `m.status = "Base branch set to: " + branch + " (warning: failed to save)"` | N/A | Change base branch - set but save failed | Warning |
| 1023 | `m.status = "Base branch set to: " + branch + " (saved)"` | N/A | Change base branch - set and saved | Success |
| 1026 | `m.status = "Base branch set to: " + branch` | N/A | Change base branch - set only | Success |
| 1075 | `m.status = "Commit subject cannot be empty"` | N/A | Commit modal - empty subject validation | Error |
| 1081 | `m.status = "Creating commit..."` | N/A | Commit modal - commit creation initiated | Info |
| 1119 | `m.status = "Failed to save editor preference"` | N/A | Editor select modal - save failed | Error |
| 1121 | `m.status = "Editor set to: " + selectedEditor` | N/A | Editor select modal - editor set | Success |
| 1221 | `m.status = "Error checking tmux config: " + err.Error()` | N/A | Tmux config modal - config check failed | Error |
| 1232 | `m.status = "Failed to update tmux config: " + err.Error()` | N/A | Tmux config modal - update failed | Error |
| 1234 | `m.status = "gcool tmux config updated! New tmux sessions will use the updated config."` | N/A | Tmux config modal - update success | Success |
| 1239 | `m.status = "Failed to remove tmux config: " + err.Error()` | N/A | Tmux config modal - remove failed | Error |
| 1241 | `m.status = "gcool tmux config removed. New tmux sessions will use your default config."` | N/A | Tmux config modal - remove success | Success |
| 1252 | `m.status = "Failed to add tmux config: " + err.Error()` | N/A | Tmux config modal - install failed | Error |
| 1254 | `m.status = "gcool tmux config installed! New tmux sessions will use this config."` | N/A | Tmux config modal - install success | Success |

---

## PART 2: m.err Assignments with Context

| Line | Assignment | Handler | Context |
|------|-----------|---------|---------|
| 34 | `m.err = msg.err` | worktreesLoadedMsg | Failed to load worktrees - auto-clears in 4s |
| 42 | `m.err = nil` | worktreesLoadedMsg success | Clear error on successful load |
| 74 | `m.err = msg.err` | branchesLoadedMsg | Failed to load branches - auto-clears in 4s |
| 82 | `m.err = nil` | branchesLoadedMsg success | Clear error on successful load |
| 88 | `m.err = msg.err` | worktreeCreatedMsg | Failed to create worktree - auto-clears in 4s |
| 116 | `m.err = msg.err` | worktreeDeletedMsg | Failed to delete worktree - auto-clears in 4s |
| 137 | `m.err = msg.err` | branchRenamedMsg | Failed to rename branch - auto-clears in 4s |
| 155 | `m.err = msg.err` | branchCheckedOutMsg | Failed to checkout branch - auto-clears in 4s |
| 186 | `m.err = msg.err` | editorOpenedMsg | Failed to open editor - auto-clears in 4s |
| 194 | `m.err = nil` | editorOpenedMsg success | Clear error on successful open |
| 203 | `m.err = msg.err` | prCreatedMsg | Failed to create PR - auto-clears in 4s |
| 211 | `m.err = nil` | prCreatedMsg success | Clear error on successful creation |
| 220 | `m.err = msg.err` | commitCreatedMsg | Failed to create commit - auto-clears in 4s |
| 237 | `m.err = nil` | commitCreatedMsg success | Clear error on successful creation |
| 251 | `m.err = msg.err` | branchPulledMsg | Failed to pull from base branch - auto-clears in 5s |
| 264 | `m.err = nil` | branchPulledMsg success | Clear error on successful pull |
| 278 | `m.err = msg.err` | refreshWithPullMsg | Failed to refresh - auto-clears in 5s |
| 287 | `m.err = nil` | refreshWithPullMsg success | Clear error on successful refresh |
| 172 | `m.err = nil` | clearErrorMsg | Explicit error clear after auto-clear timeout |
| 541 | `m.err = err` | 'C' keybinding | Failed to check for uncommitted changes - auto-clears in 3s |
| 1118 | `m.err = err` | Editor select modal | Failed to save editor preference - no auto-clear |

---

## PART 3: Auto-Clear Mechanisms

### Pattern 1: Error Auto-Clear (4 seconds)
Used for async operation failures:
- worktreesLoadedMsg error (line 37-39)
- branchesLoadedMsg error (line 77-79)
- worktreeCreatedMsg error (line 91-93)
- worktreeDeletedMsg error (line 119-121)
- branchRenamedMsg error (line 140-142)
- branchCheckedOutMsg error (line 158-160)
- editorOpenedMsg error (line 189-191)
- prCreatedMsg error (line 206-208)
- commitCreatedMsg error (line 223-225)

**Code Pattern:**
```go
return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
})
```

### Pattern 2: Success Auto-Clear (2-3 seconds)
Used for successful async operations:
- worktreeCreatedMsg success (line 105-110): 2s
- editorOpenedMsg success (line 196-198): 2s
- commitCreatedMsg success (line 241-246): 3s
- branchPulledMsg success (line 268-273): 2s
- refreshWithPullMsg success (line 291-296): 2s

**Code Pattern:**
```go
return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
}))
```

### Pattern 3: Longer Auto-Clear (5 seconds)
Used for operations with conflict/complexity:
- prCreatedMsg success (line 213-215): 5s
- branchPulledMsg error/conflict (line 259-261): 5s
- refreshWithPullMsg error (line 281-283): 5s

### Pattern 4: Commit Check Auto-Clear (3 seconds)
Keybinding-triggered validation:
- 'C' keybinding checks (line 544-546, 551-553)

### Pattern 5: clearErrorMsg Handler
Explicit clearing on timeout (line 171-174):
```go
case clearErrorMsg:
    m.err = nil
    m.status = ""
    return m, nil
```

---

## PART 4: Message Classification for showNotification()

### Error Messages (m.err != nil + m.status = "Failed to...")
**Auto-clear: 3-5 seconds**
- Failed to load worktrees
- Failed to load branches
- Failed to create worktree
- Failed to delete worktree
- Failed to rename branch
- Failed to checkout branch
- Failed to open editor: {error}
- Failed to create PR: {error}
- Failed to create commit: {error}
- Failed to pull from base branch: {error}
- Failed to refresh: {error}
- Failed to generate random name
- Failed to generate workspace path
- Failed to check for uncommitted changes
- Failed to attach to session
- Failed to kill session
- Failed to save editor preference
- Error checking tmux config: {error}
- Failed to update tmux config: {error}
- Failed to remove tmux config: {error}
- Failed to add tmux config: {error}
- Branch name is required
- Branch name cannot be empty
- Branch name cannot be empty after sanitization
- Commit subject cannot be empty

### Warning Messages (Safety checks / Conflicts)
**Auto-clear: 2-5 seconds OR persistent**
- Cannot delete current worktree
- Cannot rename main branch. Only workspace branches can be renamed.
- Base branch not set. Press 'c' to set base branch
- Cannot pull on main worktree. Use 'git pull' manually.
- Press 'y' or Enter to confirm force delete
- Cannot delete: uncommitted changes. Use 'Force Delete' to proceed.
- Merge conflict! Run 'git merge --abort' in the worktree to abort.
- Base branch set to: {branch} (warning: failed to save)

### Success Messages
**Auto-clear: 2-5 seconds**
- Worktree created successfully
- Worktree deleted successfully
- Branch renamed successfully
- Branch checked out successfully
- Opened in editor
- Draft PR created: {url}
- Commit created: {hash}
- Commit created successfully
- Successfully pulled changes from base branch
- Base branch set to: {branch} (saved)
- Base branch set to: {branch}
- Editor set to: {editor}
- Session killed
- gcool tmux config updated! New tmux sessions will use the updated config.
- gcool tmux config removed. New tmux sessions will use your default config.
- gcool tmux config installed! New tmux sessions will use this config.

### Info Messages (Operational status)
**No auto-clear (persists until next action)**
- Pulling latest commits and refreshing...
- Creating worktree with branch: {name}
- Opening in editor...
- Worktree is already up-to-date with base branch
- Pulling changes from base branch...
- Creating draft PR...
- Nothing to commit - no uncommitted changes in {branch}
- Branch name unchanged
- Renaming branch to '{name}'...
- Checking out branch: {branch}
- Creating commit...

### Generic Messages (Pass-through via statusMsg)
- User-defined status (via statusMsg type)

---

## PART 5: Special Cases and Patterns

### 1. Dual Async Operations (Batch)
Some operations reload data + auto-clear:
```go
tea.Batch(
    cmd,                                      // Reload worktrees
    tea.Tick(duration*time.Second, ...)      // Auto-clear after
)
```
- Line 105-110: Create worktree → reload + 2s clear
- Line 241-246: Commit created → reload + 3s clear
- Line 268-273: Pull success → reload + 2s clear
- Line 291-296: Refresh success → reload + 2s clear

### 2. Sequential Operations (Sequence)
Used when order matters:
```go
tea.Sequence(cmd, tea.Tick(...))
```
- Line 241-246: Commit created (alternative to Batch)
- Line 268-273: Pull success (alternative to Batch)
- Line 291-296: Refresh success (alternative to Batch)

### 3. Persistent Status (Modal Confirmations)
Some status messages persist across modal interactions:
- Line 844: Force delete confirmation
- Line 852: Delete safety check
- Line 885: General force delete flow

### 4. Status with Error Tracking
Both m.status AND m.err set simultaneously:
- All async operation failures
- Some keybinding validations (line 541-542)
- Modal-level validation (line 1118)

### 5. Dynamic Status Messages
Built from runtime data:
- Line 233: Commit hash display (`"Commit created: " + hashDisplay`)
- Line 286: Refresh status (`buildRefreshStatusMessage(msg)`)
- Line 1021: Base branch with save failure note
- Line 1284: Branch details in refresh message

### 6. Status Message Patterns in Keybindings
Messages set directly without async operation:
- Validation errors (empty fields)
- Safety checks (current worktree, main branch)
- Operation initiation ("Opening in editor...")
- Status checks ("Already up-to-date")

---

## PART 6: Replacement Strategy for showNotification()

### Messages Needing Duration Specification

**Short Duration (2 seconds):**
- Opened in editor
- Worktree created successfully
- Successfully pulled changes from base branch
- (Successful operations that reload data)

**Medium Duration (3 seconds):**
- Commit created (with hash)
- Commit created successfully

**Long Duration (4-5 seconds):**
- All "Failed to..." errors
- All warnings with user action required
- Merge conflict messages
- PR creation confirmation

**Persistent (Until Next Action):**
- All info/operational status messages
- Validation errors
- Safety checks
- Status checks

### Suggested showNotification() Signature
```go
func (m *Model) showNotification(message string, notificationType NotificationType, autoClearAfter *time.Duration)
```

Or with helper methods:
```go
func (m *Model) showSuccessNotification(message string, duration time.Duration)
func (m *Model) showErrorNotification(message string, duration time.Duration)
func (m *Model) showWarningNotification(message string, duration time.Duration)
func (m *Model) showInfoNotification(message string) // No auto-clear
```

