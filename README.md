btrz-reporter - aws memory reporter
=====================
What
----------
This is a memory reporting program. AWS metrics are not showing memory usage, this utility solves this issue.
Amazon does offer perl scripts, but these didn't work well for us.

Works on:
----------------
Ubuntu servers only, versions 14 and 16 were tested. This **will not** work on Windows.

Building
------------
You'll need `go` to build it. `make` will make you life easier, just run `make test` to test(and build) and `make run` to run.

Settings up
-------------
Any server that you use this utility on, need to have reporting permissions to AWS cloudwatch.
