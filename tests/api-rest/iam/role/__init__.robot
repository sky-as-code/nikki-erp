*** Settings ***
Documentation     IAM role suite. Covers the role-side assignment endpoints
...               (POST /v1/iam/roles/:role_id/manage-users and .../manage-groups), which
...               write the same junction rows as the principal-side endpoints
...               (POST /v1/iam/users/:user_id/roles, POST /v1/iam/groups/:group_id/roles),
...               and the entitlement delta of PATCH /v1/iam/roles/:id.
...
...               Teardown removes every user, group and role the suite created.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
Force Tags        iam    role
