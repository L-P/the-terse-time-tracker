# The Terse Time Tracker
Or `tt` for short.

Project state: usable but WIP. I will implement what I need when I need it.

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

# Show a complete timesheet:
$ tt -report | tail -n14
Week #17 from 2021-04-26 to 2021-05-02
  08:14    08:53    09:19    08:48    08:57  
  17:16    16:53    18:15    17:10    16:53  
  08h17m   07h21m   08h26m   07h53m   07h16m 
 +00h29m  -00h26m  +00h38m  +00h05m  -00h31m  +00h15m  (+00h22m)
   Mon.     Tue.     Wed.     Thu.     Fri.    Total

Week #18 from 2021-05-03 to 2021-05-09
  09:14                                      
  16:44                                      
  07h29m                                     
 -00h18m                                      -00h18m  (+00h03m)
   Mon.     Tue.     Wed.     Thu.     Fri.    Total
```
