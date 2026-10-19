*** Settings ***
Documentation     Job Scheduler Job resource suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the job under test) -> UPDATE -> GET ->
...               EXISTS -> SEARCH -> DELETE (cleanup, always last).
...
...               There is deliberately no 06_archive.robot. A scheduled job is not
...               archivable: the schema declines core.basemodel.archivable_model and
...               0002002_jobscheduler_iam.sql seeds no set_archived action, so the
...               endpoint does not exist. 08_delete.robot asserts that absence instead,
...               which documents the decision as a test rather than as a missing file.
...
...               The suite skips against the hosted coreplatform app, which deliberately does not
...               load the jobscheduler module (JOBSCHEDULER_SERVED=False in its environment file).
Resource          resources/jobscheduler.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Skip Unless Jobscheduler Is Served
Suite Teardown    Delete Jobscheduler Seed Data
