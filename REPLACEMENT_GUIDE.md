# Systematic Replacement Guide: m.status and m.err to showNotification()

## Quick Reference

**Total lines to replace: 73 assignments across 59 unique message instances**

- m.status assignments: 56 (38 unique messages + 18 clears/dynamic)
- m.err assignments: 21 (18 error sets + 3 clears)

---

## Replacement Priority

### PHASE 1: Async Operation Handlers (Easy - 20 replacements)
All in the Update() switch statement, lines 32-297

1. **worktreesLoadedMsg** (lines 34-35, 42, 43)
2. **branchesLoadedMsg** (lines 74-75, 82)
3. **worktreeCreatedMsg** (lines 88, 89, 95)
4. **worktreeDeletedMsg** (lines 116, 117, 123)
5. **branchRenamedMsg** (lines 137, 138, 144)
6. **branchCheckedOutMsg** (lines 155, 156, 162)
7. **editorOpenedMsg** (lines 186, 187, 193, 194)
8. **prCreatedMsg** (lines 203, 204, 210, 211)
9. **commitCreatedMsg** (lines 220, 221, 233, 235, 237)
10. **branchPulledMsg** (lines 251, 254, 256, 263, 264)
11. **refreshWithPullMsg** (lines 278, 279, 286, 287)

**Pattern:** All have auto-clear via tea.Tick with known durations

### PHASE 2: Keybinding Handlers (Medium - 25 replacements)
In handleMainInput() function, lines 321-570

1. **'r' refresh** (line 345)
2. **'n' create with random** (lines 352, 359, 363)
3. **'d' delete** (lines 394, 403)
4. **'R' rename** (line 427)
5. **'o' open editor** (line 467)
6. **'p' pull from base** (lines 508, 514, 520, 524)
7. **'P' create PR** (line 531)
8. **'C' commit** (lines 541, 542, 549)

**Pattern:** Validation/safety checks - mostly persistent or short duration (3s)

### PHASE 3: Modal Handlers (Medium - 20 replacements)
Lines 759-1267

1. **handleCreateModalInput** (lines 794, 801)
2. **handleDeleteModalInput** (lines 844, 852)
3. **handleBranchSelectModalInput** (line 895)
4. **handleCheckoutBranchModalInput** (line 907)
5. **handleSessionListModalInput** (lines 925, 937, 939)
6. **handleRenameModalInput** (lines 972, 979, 985, 991)
7. **handleChangeBaseBranchModalInput** (lines 1021, 1023, 1026)
8. **handleCommitModalInput** (lines 1075, 1081)
9. **handleEditorSelectModalInput** (lines 1118, 1119, 1121)
10. **handleTmuxConfigModalInput** (lines 1221, 1232, 1234, 1239, 1241, 1252, 1254)

**Pattern:** Mostly synchronous operations or modal state confirmations

### PHASE 4: Special Cases (Small - 8 replacements)
1. **clearErrorMsg handler** (lines 172, 173)
2. **statusMsg handler** (line 177)
3. **Refresh status message** (line 286 - dynamic)
4. **Build refresh status** (line 1270-1305)

---

## Message Durations by Type

### Auto-Clear 2 seconds (Short)
- "Opened in editor"
- "Worktree created successfully"
- "Successfully pulled changes from base branch"
- (All success operations that reload data)

### Auto-Clear 3 seconds (Medium)
- "Commit created: {hash}"
- "Commit created successfully"
- "Failed to check for uncommitted changes: {error}"
- "Nothing to commit - no uncommitted changes in {branch}"

### Auto-Clear 4 seconds (Standard Error)
- "Failed to load worktrees"
- "Failed to load branches"
- "Failed to create worktree"
- "Failed to delete worktree"
- "Failed to rename branch"
- "Failed to checkout branch"
- "Failed to open editor: {error}"
- "Failed to create PR: {error}"
- "Failed to create commit: {error}"

### Auto-Clear 5 seconds (Conflict/Complex)
- "Draft PR created: {url}"
- "Merge conflict! Run 'git merge --abort'..."
- "Failed to pull from base branch: {error}"
- "Failed to refresh: {error}"

### Persistent (No Auto-Clear)
- All validation errors
- All safety checks
- All operation initiation messages ("Creating...", "Opening...", "Pulling...")
- All status checks
- All modal-level messages
- Base branch configuration messages
- Tmux configuration messages

---

## Implementation Example

### Before (Current Code):
```go
case commitCreatedMsg:
    if msg.err != nil {
        m.err = msg.err
        m.status = "Failed to create commit: " + msg.err.Error()
        return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
            return clearErrorMsg{}
        })
    } else {
        if msg.commitHash != "" {
            hashDisplay := msg.commitHash
            if len(msg.commitHash) > 8 {
                hashDisplay = msg.commitHash[:8]
            }
            m.status = "Commit created: " + hashDisplay
        } else {
            m.status = "Commit created successfully"
        }
        m.err = nil
        cmd = m.loadWorktrees
        return m, tea.Sequence(
            cmd,
            tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
                return clearErrorMsg{}
            }),
        )
    }
```

### After (With showNotification):
```go
case commitCreatedMsg:
    if msg.err != nil {
        m.showErrorNotification("Failed to create commit: " + msg.err.Error(), 4*time.Second)
        return m, nil
    } else {
        if msg.commitHash != "" {
            hashDisplay := msg.commitHash
            if len(msg.commitHash) > 8 {
                hashDisplay = msg.commitHash[:8]
            }
            m.showSuccessNotification("Commit created: " + hashDisplay, 3*time.Second)
        } else {
            m.showSuccessNotification("Commit created successfully", 3*time.Second)
        }
        cmd = m.loadWorktrees
        return m, cmd
    }
```

---

## Notification Type Reference

```go
type NotificationType int

const (
    NotificationSuccess NotificationType = iota  // Green
    NotificationError                           // Red
    NotificationWarning                         // Yellow
    NotificationInfo                            // Blue/Cyan
)
```

### When to use each type:

**NotificationError:**
- "Failed to..." messages
- Technical errors with details
- Validation failures (empty fields, invalid input)
- Operation failures (attach, kill, save)
- Line errors (1118, 541)

**NotificationWarning:**
- Safety checks ("Cannot delete...", "Cannot rename...")
- Conflicts ("Merge conflict!")
- Partial failures ("...failed to save")
- Requires user action
- Confirmation prompts

**NotificationSuccess:**
- "...created successfully" messages
- "...deleted successfully" messages
- "Opened in editor"
- "Session killed"
- Configuration saved messages
- Successful dynamic operations

**NotificationInfo:**
- Operation initiation ("Creating...", "Opening...", "Pulling...")
- Status checks ("Already up-to-date", "Nothing to commit")
- Status changes ("Branch name unchanged")
- User-actionable information

---

## Bulk Replacement Patterns

### Pattern 1: Error + Status + Auto-clear 4s
```go
m.err = msg.err
m.status = "Failed to X"
return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
})
```

Replace with:
```go
m.showErrorNotification("Failed to X", 4*time.Second)
return m, nil
```

**Affected lines:** 34-39, 74-79, 88-93, 116-121, 137-142, 155-160, 186-191, 203-208, 220-225, 251-261, 278-283

### Pattern 2: Success + Status + Auto-clear 2s + Reload
```go
m.status = "Worktree created successfully"
m.modal = noModal
cmd = m.loadWorktrees
return m, tea.Batch(
    cmd,
    tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return clearErrorMsg{}
    }),
)
```

Replace with:
```go
m.showSuccessNotification("Worktree created successfully", 2*time.Second)
m.modal = noModal
return m, m.loadWorktrees
```

**Affected lines:** 95-112, 193-198, 263-273, 286-296

### Pattern 3: Status only (Persistent)
```go
m.status = "Cannot delete current worktree"
return m, nil
```

Replace with:
```go
m.showWarningNotification("Cannot delete current worktree")
return m, nil
```

**Affected lines:** 345, 352, 359, 363, 394, 403, 427, 467, 508, 514, 520, 524, 531, 794, 801, 844, 852, 895, 907, 925, 937, 972, 979, 985, 991, 1021, 1023, 1026, 1075, 1081, 1119, 1121, 1221, 1232, 1234, 1239, 1241, 1252, 1254

### Pattern 4: Dynamic Messages
Lines with fmt.Sprintf or string concatenation:
- Line 286: buildRefreshStatusMessage(msg)
- Line 233: "Commit created: " + hashDisplay
- Line 210: "Draft PR created: " + msg.prURL
- Line 1284: Branch details in refresh message

Preserve dynamic message logic, wrap result in showNotification()

---

## Error Handling Strategy

**Current approach:**
- Both m.err and m.status are set
- m.err stores actual error object
- m.status stores formatted message
- View layer displays m.status, ignores m.err (mostly)

**New approach:**
- showNotification() encapsulates message + type
- No need to track m.err separately for display
- Keep m.err = nil assignments for clarity during transition
- Eventually can remove m.err field entirely

### Transition Path:
1. Replace m.status/m.err assignments with showNotification() calls
2. Keep calling m.err = nil where appropriate (marks intent)
3. After full replacement, remove m.err field and clearErrorMsg handler
4. Update view layer to display notification instead of m.status

---

## Testing Checklist

After replacements, verify:

- [ ] All error messages display with red color
- [ ] All warning messages display with yellow color
- [ ] All success messages display with green color
- [ ] All info messages display with blue/cyan color
- [ ] Short messages (2-3s) auto-clear quickly
- [ ] Standard messages (4s) auto-clear normally
- [ ] Persistent messages stay until next user action
- [ ] Multiple notifications queue properly
- [ ] Error details are included in messages
- [ ] Dynamic content (hashes, URLs, branches) is included

---

## Message Count Summary

| Category | Count | Type |
|----------|-------|------|
| Async Operation Errors (4s) | 9 | NotificationError |
| Async Operation Success | 6 | NotificationSuccess |
| Keybinding Validation (Persistent) | 18 | Mixed |
| Modal Operations (Persistent) | 20 | Mixed |
| Configuration Changes | 4 | NotificationSuccess/Warning |
| Session Operations | 3 | Mixed |
| Tmux Configuration | 6 | NotificationSuccess/Error |
| Special Cases (Clear, Generic) | 3 | N/A |
| **TOTAL** | **69** | **Various** |

