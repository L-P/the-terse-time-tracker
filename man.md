% TT(1) tt $VERSION
% Léo Peltier
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
    tasks tags. If no tags are given and there is a task running, the new tasks
    will reuse the running task tags.

stop
:   Stops the current task timer.
