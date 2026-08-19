*** Settings ***
Documentation     IAM permission suite. Proves that the entitlement-based permission
...               system grants and revokes access correctly through the public API, and
...               that the self-service probe (POST /v1/iam/me/test-permissions) always
...               agrees with what enforcement actually does.
...
...               File order encodes the flow: CONTRACT (endpoint behaviour) ->
...               ROLE LIFECYCLE -> GROUP LIFECYCLE -> SCOPE MATRIX -> ESCALATION.
...               Every lifecycle assertion pairs the probe with a real API call, so a
...               probe that lied about a grant would be caught rather than believed.
...
...               Teardown removes every user, group, role and entitlement the suite
...               created; nothing here depends on pre-provisioned fixtures beyond the
...               administrator credentials the whole test run already uses.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
Force Tags        iam    permission    security
