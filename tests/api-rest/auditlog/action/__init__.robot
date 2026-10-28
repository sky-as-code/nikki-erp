*** Settings ***
Documentation     Audit Log suite (AUDIT-004..007). Every mutation of an engine resource is
...               recorded in the write's transaction with a diff of the changed fields, and
...               the history route answers whoever may read the target record. The target is
...               a UoM category the suite owns. File order: MUTATIONS (create, update, no-op
...               update, archive, unarchive) -> REQUEST ID -> HISTORY ACCESS -> READS ->
...               DELETE (deleted target, always last).
...
...               Covered by unit tests instead: a failed audit insert rolls the change back
...               (composable TestAuditTrailFailedRecordRollsTheWriteBack), since nothing over
...               REST can make the insert fail.
Resource          resources/auditlog.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Audit Fixtures
