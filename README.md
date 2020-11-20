# The Terse Time Tracker
Or `tt` for short.

## Usage
```sh
# Start working on a task:
$ tt start working on tt

# If the first word is not a recognized command, the entire arguments list is
# interpreted as a task description, switching to another task is just a matter
# of calling start again:
$ tt writing documentatin

# Done with the work, stop the task:
$ tt stop

# There was a typo in the last task, let's edit it:
$ tt amend writing documentation

# Until now the tasks had no tags, let's categorize them:
$ tt foobaring @acme-corp @billable

# It is assumed you want to keep the previous tags when switching tasks:
$ tt bazzing

# If you want to remove a tag, amend with a new tag list that will overwrite
# the previous one:
$ tt bazzing @acme-corp


# We're done for now.
$ tt stop
```
