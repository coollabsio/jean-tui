# Notification Refactoring - Complete Analysis Index

This directory contains the complete analysis for replacing m.status and m.err assignments with a showNotification() system in tui/update.go.

## Documents Overview

### 1. QUICK_REFERENCE.md (START HERE)
**Best for:** Quick overview before starting implementation
- 1-page cheat sheet with all key information
- 4 most common replacement patterns with before/after code
- Helper method signatures
- Migration steps (5 phases)
- 10-minute read

### 2. REPLACEMENT_GUIDE.md (DETAILED STRATEGY)
**Best for:** Understanding the systematic approach
- 4-phase replacement strategy (Easy → Medium → Complex)
- Bulk replacement patterns with context
- Message durations by type
- Implementation examples (detailed)
- Error handling strategy
- 30-minute read

### 3. STATUS_ANALYSIS.md (COMPLETE REFERENCE)
**Best for:** Finding specific context for any line
- 6 parts covering all angles
- All 69 assignments documented
- Auto-clear mechanisms explained
- Message classification system
- Special cases and patterns
- 40+ page comprehensive reference

### 4. status_assignments.csv (SPREADSHEET FORMAT)
**Best for:** Filtering and organizing data
- All assignments in CSV format
- Sortable by: line, type, category, duration
- Import into Excel/Google Sheets for sorting
- Good for tracking replacement progress

## Quick Stats

| Metric | Value |
|--------|-------|
| Total assignments to replace | 77 lines |
| m.status assignments | 56 |
| m.err assignments | 21 |
| Notification types needed | 4 |
| Estimated effort | 2-3 hours |
| Complexity | Medium |

## Notification Types

```
NotificationSuccess (Green)  - Successful operations
NotificationError (Red)      - Operation failures
NotificationWarning (Yellow) - Safety checks, conflicts
NotificationInfo (Cyan)      - Status, operation initiation
```

## Auto-Clear Durations

- 2s: Quick success operations
- 3s: Commit operations  
- 4s: Standard errors (most common)
- 5s: Complex/conflicts
- None: Persistent messages (validation, safety, status)

## Files Modified by Replacement

- **tui/update.go** - Main changes (77 lines)
- **tui/model.go** - Add helper methods
- **tui/view.go** - Update to display notifications (if needed)

## Implementation Phases

### Phase 1: Async Operations (Lines 32-297)
- 20 replacements, most straightforward
- All follow clear patterns
- Error handlers with 4s auto-clear

### Phase 2: Keybindings (Lines 321-570)
- 25 replacements, mix of persistent/timed
- Validation messages
- Safety checks

### Phase 3: Modals (Lines 759-1267)
- 20 replacements, mostly persistent
- Confirmation prompts
- Dynamic content (fmt.Sprintf)

### Phase 4: Cleanup (Scattered)
- Remove clearErrorMsg handler
- Remove tea.Tick patterns
- Optional: Remove m.err field

## Common Patterns

### Pattern A: Error + Auto-clear 4s
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

### Pattern B: Success + Reload + Auto-clear 2s
```go
// Before
m.status = "Worktree created successfully"
cmd = m.loadWorktrees
return m, tea.Batch(cmd, tea.Tick(2*time.Second, ...))

// After
m.showSuccessNotification("Worktree created successfully", 2*time.Second)
return m, m.loadWorktrees
```

### Pattern C: Persistent Status
```go
// Before
m.status = "Cannot delete current worktree"
return m, nil

// After
m.showWarningNotification("Cannot delete current worktree")
return m, nil
```

### Pattern D: Dynamic Message
```go
// Before
m.status = "Commit created: " + hashDisplay

// After
m.showSuccessNotification("Commit created: " + hashDisplay, 3*time.Second)
```

## Next Steps

1. Open QUICK_REFERENCE.md to understand the 4 patterns
2. Review the helper method signatures
3. Implement helper methods in tui/model.go
4. Start Phase 1 replacements (async handlers)
5. Test and move to Phase 2
6. Continue through Phases 3 and 4

## File Locations

All analysis files are in the root of /Users/heyandras/devel/gcool/:
- QUICK_REFERENCE.md
- REPLACEMENT_GUIDE.md
- STATUS_ANALYSIS.md
- status_assignments.csv
- NOTIFICATION_REFACTOR_INDEX.md (this file)

## Quick Links to Specific Sections

**STATUS_ANALYSIS.md:**
- [PART 1: All m.status assignments](STATUS_ANALYSIS.md#part-1-mstatus-assignments-with-context)
- [PART 2: All m.err assignments](STATUS_ANALYSIS.md#part-2-merr-assignments-with-context)
- [PART 3: Auto-clear mechanisms](STATUS_ANALYSIS.md#part-3-auto-clear-mechanisms)
- [PART 4: Message classification](STATUS_ANALYSIS.md#part-4-message-classification-for-shownotification)
- [PART 5: Special cases](STATUS_ANALYSIS.md#part-5-special-cases-and-patterns)
- [PART 6: Replacement strategy](STATUS_ANALYSIS.md#part-6-replacement-strategy-for-shownotification)

**REPLACEMENT_GUIDE.md:**
- [Phase 1: Async operations](REPLACEMENT_GUIDE.md#phase-1-async-operation-handlers-easy---20-replacements)
- [Phase 2: Keybindings](REPLACEMENT_GUIDE.md#phase-2-keybinding-handlers-medium---25-replacements)
- [Phase 3: Modals](REPLACEMENT_GUIDE.md#phase-3-modal-handlers-medium---20-replacements)
- [Bulk replacement patterns](REPLACEMENT_GUIDE.md#bulk-replacement-patterns)
- [Testing checklist](REPLACEMENT_GUIDE.md#testing-checklist)

## Questions?

Each document has extensive context and examples. Use:
- QUICK_REFERENCE.md for "how do I start?"
- REPLACEMENT_GUIDE.md for "what's the strategy?"
- STATUS_ANALYSIS.md for "what about line 234?"
- status_assignments.csv for filtering and organizing

---

Last updated: 2025-10-29
Total lines analyzed: 77
Unique messages: 69
Patterns identified: 5
Ready for implementation: YES
