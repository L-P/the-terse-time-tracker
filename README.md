# The Terse Time Tracker
Or `tt` for short.

## Usage
```sh
# Start working on a task:
$ tt start working on tt
Created task: "start working on tt"

# If no flag is given, the entire arguments list is interpreted as a task
# description, switching to another task is just a matter
# of calling start again:
$ tt writing documentation
Stopped task that had been running for 5s: "start working on tt"
Created task: "writing documentation"

# Done with the work, stop the task:
$ tt -stop
Stopped task that had been running for 5s: "writing documentation"

# The last task had no tags, let's create a task with tags:
$ tt foobaring @acme-corp @billable
Created task: "foobaring"
With tags: "@acme-corp @billable"

# It is assumed you want to keep the previous tags when switching tasks:
$ tt bazzing
Stopped task that had been running for 5s: "foobaring"
Created task: "bazzing"
With tags: "@acme-corp @billable"

# If you want to remove a tag, amend with a new tag list that will overwrite
# the previous one:
$ tt bazzing @acme-corp
Replaced tags from current task: @acme-corp

# You can also omit the task description to update tags:
$ tt @world-company
Replaced tags from current task: @world-company

# We're done for now.
$ tt -stop
Stopped task that had been running for 24s: "bazzing"
```
