*** Settings ***
Documentation     Job Scheduler Execution resource suite.
...
...               There are only three files, and that is the contract rather than an
...               omission: executions and attempts are written by the scheduling engine
...               and exposed read-only. History an API could create, edit or delete
...               could not be trusted to say what actually ran, so the seeds grant only
...               the "read" action on both resources and no write route exists.
...
...               The suite therefore reads history it did not create. Where the database
...               has none, tests skip rather than invent a fixture: skipping says the
...               contract was not exercised, where a hand-written row would imply it passed.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
