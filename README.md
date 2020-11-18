# The Terse Time Tracker
Or `tt` for short.

## Usage
```sh
# Start working on a task, the start command is implied:
$ tt working on tt

# If the first word is not a recognized command, the entire arguments list is
# interpreted as a task description, switching to another task is just a matter
# of calling tt again:
$ tt writing documentatin

# Done with the work, stop the task:
$ tt done

# There was a typo in the last task, let's edit it:
$ tt amend writing documentation

# Until now the tasks have been tagged with the default tag,
# add something more specific:
$ tt foobaring @acme-corp @billable

# It is assumed you want to keep the previous tags when switching tasks over:
$ tt bazzing

# If you want to remove a tag, amend with a new tag list that will override the
previous one:
$ tt amend @acme-corp

# Since we're editing the current running task, would could as well reuse the
previous command without having a new task created:
$ tt bazzing @acme-corp


# We're done for now.
$ tt done
```
