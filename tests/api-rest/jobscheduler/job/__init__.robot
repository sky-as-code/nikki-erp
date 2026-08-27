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
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Jobscheduler Seed Data
