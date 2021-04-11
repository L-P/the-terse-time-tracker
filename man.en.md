% TT(1) tt | $VERSION
% Written by Léo Peltier
% $DATE

# NAME
tt - The Terse Time Tracker

# SYNOPSIS
tt [-start] description  
tt *option*

# DESCRIPTION
The Terse Time Tracker tracks time spent on tasks.

# OPTIONS
If no option is given, *start* is implied. If no option and no descriptin is
given, tt will output the current running task or exit with code 1 if there is
no current task.

*-start*
:   Either creates a new task and stops the current one, or edits the current
    task tags. If no tags are given and there is a different task running, the
    new task will reuse the running task tags.

*-stop*
:   Stops the current task timer.

*-ui*
:   Starts the TUI that allows editing past entries and changing the
    configuration.

*-v*
:   Displays the version and exits.

# DATA
The Terse Time Tracker stores all of its data in a SQLite database named
`the-terse-time-tracker.db` in your default user configuration directory.

The configuration is stored in the same database and can be accessed through
the TUI by pressing F2.
