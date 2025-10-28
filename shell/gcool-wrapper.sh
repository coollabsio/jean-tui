#!/usr/bin/env bash
# Shell wrapper for gcool to enable directory switching and tmux session management
# Source this file in your shell rc file

# Bash/Zsh wrapper
gcool() {
    # Loop until user explicitly quits gcool (not just detaches from tmux)
    while true; do
        # Save current PATH to restore it later
        local saved_path="$PATH"

        # Create a temp file for communication
        local temp_file=$(mktemp)

        # Set environment variable so gcool knows to write to file
        GCOOL_SWITCH_FILE="$temp_file" command gcool "$@"
        local exit_code=$?

        # Restore PATH if it got corrupted
        if [ -z "$PATH" ] || [ "$PATH" != "$saved_path" ]; then
            export PATH="$saved_path"
        fi

        # Check if switch info was written
        if [ -f "$temp_file" ] && [ -s "$temp_file" ]; then
        # Read the switch info: path|branch|auto-claude|terminal-only
        local switch_info=$(cat "$temp_file")
        # Only remove if it's in /tmp (safety check)
        if [[ "$temp_file" == /tmp/* ]] || [[ "$temp_file" == /var/folders/* ]]; then
            rm "$temp_file"
        fi

        # Parse the info (using worktree_path instead of path to avoid PATH conflict)
        IFS='|' read -r worktree_path branch auto_claude terminal_only <<< "$switch_info"

        # Check if we got valid data (has at least two pipes)
        if [[ "$switch_info" == *"|"*"|"* ]]; then
            # Check if tmux is available
            if ! command -v tmux >/dev/null 2>&1; then
                # No tmux, just cd
                cd "$worktree_path" || return
                echo "Switched to worktree: $branch (no tmux)"
                return
            fi

            # Sanitize branch name for tmux session
            local session_name="gcool-${branch//[^a-zA-Z0-9\-_]/-}"
            session_name="${session_name//--/-}"
            session_name="${session_name#-}"
            session_name="${session_name%-}"

            # If terminal-only, append -terminal suffix
            if [ "$terminal_only" = "true" ]; then
                session_name="${session_name}-terminal"
            fi

            # Check if already in a tmux session
            if [ -n "$TMUX" ]; then
                # Already in tmux, just cd
                cd "$worktree_path" || return
                echo "Switched to worktree: $branch"
                echo "Note: Already in tmux. Session: $session_name would be available outside tmux."
                return
            fi

            # Check if session exists (use exact matching with =)
            if tmux has-session -t "=$session_name" 2>/dev/null; then
                # Attach to existing session
                tmux attach-session -t "$session_name"
                # After detaching, loop back to gcool
                continue
            else
                # Create new session in detached mode first
                # Terminal-only sessions always use shell, never Claude
                if [ "$terminal_only" = "true" ]; then
                    # Always start with shell for terminal sessions
                    tmux new-session -d -s "$session_name" -c "$worktree_path"
                elif [ "$auto_claude" = "true" ]; then
                    # Check if claude is available
                    if command -v claude >/dev/null 2>&1; then
                        # Create detached session with claude
                        tmux new-session -d -s "$session_name" -c "$worktree_path" claude
                    else
                        # Fallback: create detached session with shell and show message
                        tmux new-session -d -s "$session_name" -c "$worktree_path" \; \
                            send-keys "echo 'Note: Claude CLI not found. Install it or use --no-claude flag.'" C-m \; \
                            send-keys "echo 'You are in: $worktree_path'" C-m
                    fi
                else
                    # Create detached session with shell
                    tmux new-session -d -s "$session_name" -c "$worktree_path"
                fi

                # Now attach to the session
                tmux attach-session -t "$session_name"
                # After detaching from newly created session, loop back to gcool
                continue
            fi
        else
            return 1
        fi
        else
            # No switch file, user quit gcool without selecting a worktree
            # Only remove if it's in /tmp (safety check)
            if [[ "$temp_file" == /tmp/* ]] || [[ "$temp_file" == /var/folders/* ]]; then
                rm -f "$temp_file"
            fi
            # Exit the loop
            return $exit_code
        fi
    done
}
