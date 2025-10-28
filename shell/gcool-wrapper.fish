#!/usr/bin/env fish
# Shell wrapper for gcool to enable directory switching and tmux session management (Fish shell)
# Source this file in your config.fish

function gcool
    # Loop until user explicitly quits gcool (not just detaches from tmux)
    while true
        # Create a temp file for communication
        set temp_file (mktemp)

        # Set environment variable so gcool knows to write to file
        set -x GCOOL_SWITCH_FILE $temp_file
        command gcool $argv
        set exit_code $status

        # Check if switch info was written
        if test -f "$temp_file" -a -s "$temp_file"
            # Read the switch info: path|branch|auto-claude|terminal-only
            set switch_info (cat $temp_file)
            rm $temp_file

            # Parse the info (using worktree_path instead of path to avoid PATH conflict)
            set parts (string split '|' $switch_info)

            # Check if we got valid data (has at least 3 parts)
            if test (count $parts) -ge 3
                set worktree_path $parts[1]
                set branch $parts[2]
                set auto_claude $parts[3]
                set terminal_only "false"
                if test (count $parts) -ge 4
                    set terminal_only $parts[4]
                end

                # Check if tmux is available
                if not command -v tmux &> /dev/null
                    # No tmux, just cd
                    cd $worktree_path
                    echo "Switched to worktree: $branch (no tmux)"
                    return
                end

                # Sanitize branch name for tmux session
                set session_name "gcool-"(string replace -ra '[^a-zA-Z0-9\-_]' '-' $branch)
                set session_name (string replace -ra '--+' '-' $session_name)
                set session_name (string trim -c '-' $session_name)

                # If terminal-only, append -terminal suffix
                if test "$terminal_only" = "true"
                    set session_name "$session_name-terminal"
                end

                # Check if already in a tmux session
                if test -n "$TMUX"
                    # Already in tmux, just cd
                    cd $worktree_path
                    echo "Switched to worktree: $branch"
                    echo "Note: Already in tmux. Session: $session_name would be available outside tmux."
                    return
                end

                # Check if session exists (use exact matching with =)
                if tmux has-session -t "=$session_name" 2>/dev/null
                    # Attach to existing session
                    tmux attach-session -t "$session_name"
                    # After detaching, continue the loop to allow switching sessions
                    continue
                else
                    # Create new session
                    # Terminal-only sessions always use shell, never Claude
                    if test "$terminal_only" = "true"
                        # Always start with shell for terminal sessions
                        tmux new-session -s "$session_name" -c "$worktree_path"
                    else if test "$auto_claude" = "true"
                        # Check if claude is available
                        if command -v claude &> /dev/null
                            # Start with claude
                            tmux new-session -s "$session_name" -c "$worktree_path" claude
                        else
                            # Fallback: start with shell and show message
                            tmux new-session -s "$session_name" -c "$worktree_path"
                        end
                    else
                        # Start with shell
                        tmux new-session -s "$session_name" -c "$worktree_path"
                    end
                    # After creating and attaching, continue loop to allow switching
                    continue
                end
            end
        else
            # No switch file, just clean up
            rm -f $temp_file
            # Exit the loop
            return $exit_code
        end
    end
end
