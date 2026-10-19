*** Settings ***
Documentation     IAM role suite. Covers the role-side assignment endpoints
...               (POST /v1/iam/iam_role/:id/manage_users and .../manage_groups), which
...               write the same junction rows as the principal-side endpoints
...               (POST /v1/iam/iam_user/:id/manage_roles, POST /v1/iam/iam_group/:id/manage_roles),
...               and the entitlement delta of PATCH /v1/iam/iam_role/:id.
...
...               Teardown removes every user, group and role the suite created.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
Force Tags        iam    role
