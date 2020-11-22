% TT(1) tt $VERSION
% Written by Léo Peltier
% $DATE

# NAME
tt - The Terse Time Tracker

# SYNOPSIS
tt [*command*] description

# DESCRIPTION
The Terse Time Tracker tracks time spent on tasks and creates reports.

# COMMANDS
If no recognized *command* is specified **start** is implied.

start
:   Either creates a new task and stops the current one, or edits the current
    task tags. If no tags are given and there is a different task running, the
    new task will reuse the running task tags.

stop
:   Stops the current task timer.

ui
:   Starts the TUI that allow editing past entries.

# DATA
The Terse Time Tracker stores all of its data in a SQLite database named
`the-terse-time-tracker.db` in your default user configuration directory.
