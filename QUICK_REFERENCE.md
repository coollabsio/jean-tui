# Quick Reference: m.status/m.err Replacement Summary

## Files to Review
- `/Users/heyandras/devel/gcool/STATUS_ANALYSIS.md` - Complete detailed analysis with all context
- `/Users/heyandras/devel/gcool/REPLACEMENT_GUIDE.md` - Systematic replacement strategy
- `/Users/heyandras/devel/gcool/status_assignments.csv` - All assignments in table format

## Key Statistics
- **Total m.status assignments:** 56
- **Total m.err assignments:** 21
- **Total lines to modify:** ~77 across tui/update.go
- **Notification types needed:** 4 (Success, Error, Warning, Info)

## The Four Notification Types

```go
const (
    NotificationSuccess  // Green - "...created", "...successfully"
    NotificationError    // Red - "Failed to..."
    NotificationWarning  // Yellow - Safety checks, conflicts
    NotificationInfo     // Cyan - Status, operation initiation
)
```

## Auto-Clear Durations

| Duration | Use Case | Examples |
|----------|----------|----------|
| None | Persistent info/validation | "Creating...", "Base branch not set"  |
| 2s | Quick success | "Opened in editor", "Worktree created" |
| 3s | Commit operations | "Commit created: {hash}" |
| 4s | Standard errors | "Failed to load worktrees" |
| 5s | Complex/conflicts | "Merge conflict!", "Draft PR created" |

## Most Common Patterns (Copy-Paste Ready)

### Pattern A: Error with auto-clear (4s)
```go
// Before
m.err = msg.err
m.status = "Failed to X"
return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
})

// After
m.showErrorNotification("Failed to X", 4*time.Second)
return m, nil
```

### Pattern B: Success with reload + auto-clear (2s)
```go
// Before
m.status = "Worktree created successfully"
m.modal = noModal
cmd = m.loadWorktrees
return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
}))

// After
m.showSuccessNotification("Worktree created successfully", 2*time.Second)
m.modal = noModal
return m, m.loadWorktrees
```

### Pattern C: Persistent info (no auto-clear)
```go
// Before
m.status = "Opening in editor..."
return m, m.openInEditor(wt.Path)

// After
m.showInfoNotification("Opening in editor...")
return m, m.openInEditor(wt.Path)
```

### Pattern D: Warning (persistent)
```go
// Before
m.status = "Cannot delete current worktree"
return m, nil

// After
m.showWarningNotification("Cannot delete current worktree")
return m, nil
```

## Replacement by Section

### Update() handlers (lines 32-297) - 20 replacements
- All async operation error handlers: errors set + status + tea.Tick(4s)
- Straightforward 1:1 replacements
- Remove clearErrorMsg patterns

### handleMainInput() (lines 321-570) - 25 replacements
- Keybinding validation messages
- Mix of persistent and 3s auto-clear messages
- Mostly simple status assignments

### Modal handlers (lines 759-1267) - 20 replacements
- Modal confirmation/validation messages
- Mostly persistent (until user acts)
- Some dynamic messages (fmt.Sprintf)

### Special cases - 8 replacements
- clearErrorMsg handler removal
- statusMsg pass-through handling
- Dynamic message building

## Helper Methods to Implement

```go
func (m *Model) showErrorNotification(message string, duration time.Duration)
func (m *Model) showSuccessNotification(message string, duration time.Duration)
func (m *Model) showWarningNotification(message string, duration time.Duration)
func (m *Model) showInfoNotification(message string) // No duration
```

OR single unified method:
```go
func (m *Model) showNotification(message string, notificationType NotificationType, autoClearAfter *time.Duration)
```

## Lines to Remove After Replacement

1. **Line 172-174:** clearErrorMsg handler case
2. **All tea.Tick(Xs, func(t time.Time) tea.Msg { return clearErrorMsg{} })** patterns
3. **m.err = nil** assignments (optional, for clarity can keep some)
4. Eventually: m.err field from Model struct

## Before & After Examples

### Example 1: Basic Error
**Before (lines 34-39):**
```go
if msg.err != nil {
    m.err = msg.err
    m.status = "Failed to load worktrees"
    return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
        return clearErrorMsg{}
    })
}
```

**After:**
```go
if msg.err != nil {
    m.showErrorNotification("Failed to load worktrees", 4*time.Second)
    return m, nil
}
```

### Example 2: Success with Reload
**Before (lines 95-110):**
```go
m.status = "Worktree created successfully"
m.modal = noModal
m.lastCreatedBranch = msg.branch
cmd = m.loadWorktrees
return m, tea.Batch(
    cmd,
    tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return clearErrorMsg{}
    }),
)
```

**After:**
```go
m.showSuccessNotification("Worktree created successfully", 2*time.Second)
m.modal = noModal
m.lastCreatedBranch = msg.branch
return m, m.loadWorktrees
```

### Example 3: Persistent Warning
**Before (line 403):**
```go
m.status = "Cannot delete current worktree"
```

**After:**
```go
m.showWarningNotification("Cannot delete current worktree")
```

## Error Detail Messages (Dynamic)

These preserve error details - keep the pattern:

```go
// Before
m.status = "Failed to X: " + msg.err.Error()

// After
m.showErrorNotification("Failed to X: " + msg.err.Error(), 4*time.Second)
```

Lines with this pattern:
- 187: "Failed to open editor: " + msg.err.Error()
- 204: "Failed to create PR: " + msg.err.Error()
- 221: "Failed to create commit: " + msg.err.Error()
- 256: "Failed to pull from base branch: " + msg.err.Error()
- 279: "Failed to refresh: " + msg.err.Error()
- 1221: "Error checking tmux config: " + err.Error()
- 1232: "Failed to update tmux config: " + err.Error()
- 1239: "Failed to remove tmux config: " + err.Error()
- 1252: "Failed to add tmux config: " + err.Error()

## Dynamic Content Messages

Keep the build logic, wrap result in notification:

```go
// Line 233 - Commit hash
hashDisplay := msg.commitHash
if len(msg.commitHash) > 8 {
    hashDisplay = msg.commitHash[:8]
}
m.showSuccessNotification("Commit created: " + hashDisplay, 3*time.Second)

// Line 210 - PR URL
m.showSuccessNotification("Draft PR created: " + msg.prURL, 5*time.Second)

// Line 286 - Refresh status
m.showSuccessNotification(buildRefreshStatusMessage(msg), 2*time.Second)
```

## Testing Notes

After implementing showNotification():

1. **Error messages** should display in red
2. **Warning messages** should display in yellow
3. **Success messages** should display in green
4. **Info messages** should display in cyan/blue
5. **Auto-clear timing** should match specified durations
6. **Persistent messages** should remain until next user action
7. **Multiple notifications** should queue/stack properly
8. **Error details** should be included in final message

## Migration Strategy

### Step 1: Add helper methods to Model
Implement the 4 notification helpers (or 1 unified method)

### Step 2: Replace Phase 1 (Async handlers)
Lines 32-297 - Update() switch handlers
Most straightforward, clear patterns

### Step 3: Replace Phase 2 (Keybindings)
Lines 321-570 - handleMainInput()
Mix of persistent and timed messages

### Step 4: Replace Phase 3 (Modals)
Lines 759-1267 - Modal input handlers
Mostly persistent, some validation

### Step 5: Clean up
Remove clearErrorMsg handler
Update view layer if needed
Remove m.err field assignments

## References

For detailed context on any line:
- See STATUS_ANALYSIS.md for all 69 assignments with context
- See REPLACEMENT_GUIDE.md for systematic phase-by-phase approach
- See status_assignments.csv for spreadsheet-compatible format
