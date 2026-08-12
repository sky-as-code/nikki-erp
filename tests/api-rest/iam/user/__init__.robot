*** Settings ***
Documentation     IAM User resource suite. File order (NN_ prefixes) encodes the
...               mandated flow: CREATE (saves the user under test) -> UPDATE -> GET ->
...               EXISTS -> ARCHIVE -> SEARCH -> DELETE (cleanup, always last).
...               Teardown removes only seed entities created on the fly.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Seed Data
