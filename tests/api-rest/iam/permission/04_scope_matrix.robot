*** Settings ***
Documentation     Scope semantics over the wire.
...
...               The rules under test: a wider grant satisfies a narrower requirement
...               (domain covers org, org covers the units inside it), a grant naming one
...               org or unit never answers for a different one, a bare org grant answers
...               only for a caller who is actually a member, and a grant on a parent unit
...               does NOT reach its children.
...
...               That last one is deliberate rather than an oversight: unit grants apply
...               to the unit itself so that moving a unit in the tree cannot silently
...               change anybody's access.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Scope Fixtures


*** Keywords ***
Set Up Scope Fixtures
    [Documentation]    Two orgs and a parent/child unit pair, plus a user with their own
    ...    session, so every scope combination has something concrete to name.
    Create Authorized API Session
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_user    read
    Set Suite Variable    ${SCOPE_RESOURCE_ID}    ${resource_id}
    Set Suite Variable    ${SCOPE_ACTION_ID}    ${action_id}

    ${user_id}    ${email}=    Create Probe User Session    scope_user
    Set Suite Variable    ${SCOPE_USER_ID}    ${user_id}

    ${org_a}=    Create Probe Org    Perm Scope Org A
    ${org_b}=    Create Probe Org    Perm Scope Org B
    Set Suite Variable    ${ORG_A}    ${org_a}
    Set Suite Variable    ${ORG_B}    ${org_b}

    ${parent}=    Create Probe Org Unit    Perm Scope Unit Parent    ${org_a}
    ${child}=    Create Probe Org Unit    Perm Scope Unit Child    ${org_a}    ${parent}
    Set Suite Variable    ${UNIT_PARENT}    ${parent}
    Set Suite Variable    ${UNIT_CHILD}    ${child}

Create Probe Org
    [Arguments]    ${label}
    ${name}=    Unique Display Name    ${label}
    ${slug}=    Evaluate    "${label}".lower().replace(' ', '-') + '-' + str(__import__('uuid').uuid4())[:8]
    ${resp}=    POST On Session    api    /v1/iam/organizations
    ...    json=${{ {'display_name': $name, 'slug': $slug, 'legal_name': $name} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Create Probe Org Unit
    [Arguments]    ${label}    ${org_id}    ${parent_id}=${EMPTY}
    ${name}=    Unique Display Name    ${label}
    ${payload}=    Create Dictionary    name=${name}    display_name=${name}    org_id=${org_id}
    IF    $parent_id
        Set To Dictionary    ${payload}    parent_id=${parent_id}
    END
    ${resp}=    POST On Session    api    /v1/iam/orgunits    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Grant Scoped Role
    [Documentation]    Creates and assigns a role whose single entitlement carries the
    ...    given scope, then returns the role id so the caller can revoke it again.
    [Arguments]    ${label}    ${scope}    ${scope_field}=${EMPTY}    ${scope_id}=${EMPTY}
    ${role_id}=    Create Probe Role    ${label}    ${SCOPE_ACTION_ID}    ${SCOPE_RESOURCE_ID}
    ...    scope=${scope}    scope_field=${scope_field}    scope_id=${scope_id}
    Assign Role To User    ${SCOPE_USER_ID}    ${role_id}
    RETURN    ${role_id}

Skip Unless Unit Scope Is Grantable
    [Documentation]    A unit-scoped entitlement can only exist on a resource whose
    ...    min_scope reaches orgunit. No seeded IAM resource does today — they all bottom
    ...    out at org — so these cases are skipped rather than failed: the evaluation rule
    ...    they cover is proven by the unit tests, and failing here would report a missing
    ...    seed as a broken permission system.
    ${resp}=    GET On Session    api    /v1/iam/resources
    ...    params=${{ {'graph': '{"if":["min_scope", "=", "orgunit"]}'} }}    expected_status=any
    IF    ${resp.status_code} != 200 or len($resp.json().get('items', [])) == 0
        Skip    No resource permits orgunit-scoped entitlements in this deployment
    END


*** Test Cases ***
A Domain Grant Satisfies An Org Scoped Question
    [Documentation]    Widening: domain is the widest scope, so it answers a question
    ...    asked about any particular org.
    ${role_id}=    Grant Scoped Role    Perm Scope Domain    domain
    Permission Should Be Granted    scope_user    read:iam_user:org/${ORG_A}
    Permission Should Be Granted    scope_user    read:iam_user:org/${ORG_B}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

A Domain Grant Satisfies A Unit Scoped Question
    ${role_id}=    Grant Scoped Role    Perm Scope Domain Unit    domain
    Permission Should Be Granted    scope_user    read:iam_user:orgunit/${UNIT_CHILD}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

An Org Grant Answers Only For Its Own Org
    [Documentation]    Narrowing in the other direction: naming org A must not grant
    ...    anything about org B. This is the isolation the org scope exists to provide.
    ${role_id}=    Grant Scoped Role    Perm Scope Org A    org    org_id    ${ORG_A}
    Permission Should Be Granted    scope_user    read:iam_user:org/${ORG_A}
    Permission Should Be Denied    scope_user    read:iam_user:org/${ORG_B}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

An Org Grant Satisfies A Question About A Unit Inside That Org
    [Documentation]    The org-to-unit fallback: a unit belongs to an org, so a grant over
    ...    that org covers the records inside its units.
    ${role_id}=    Grant Scoped Role    Perm Scope Org Fallback    org    org_id    ${ORG_A}
    Permission Should Be Granted    scope_user    read:iam_user:orgunit/${UNIT_CHILD}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

A Unit Grant Answers Only For Its Own Unit
    Skip Unless Unit Scope Is Grantable
    ${role_id}=    Grant Scoped Role    Perm Scope Unit Child    orgunit    org_unit_id    ${UNIT_CHILD}
    Permission Should Be Granted    scope_user    read:iam_user:orgunit/${UNIT_CHILD}
    Permission Should Be Denied    scope_user    read:iam_user:orgunit/${UNIT_PARENT}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

A Parent Unit Grant Does Not Reach A Child Unit
    [Documentation]    No downward inheritance between units. A grant on the parent must
    ...    not reach the child, so that reorganising the tree never silently widens
    ...    somebody's access — every unit grant stays exactly as auditable as it was
    ...    written.
    Skip Unless Unit Scope Is Grantable
    ${role_id}=    Grant Scoped Role    Perm Scope Unit Parent    orgunit    org_unit_id    ${UNIT_PARENT}
    Permission Should Be Granted    scope_user    read:iam_user:orgunit/${UNIT_PARENT}
    Permission Should Be Denied    scope_user    read:iam_user:orgunit/${UNIT_CHILD}
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

A Narrow Grant Does Not Satisfy A Domain Question
    [Documentation]    Widening runs one way only. Holding a permission over one org says
    ...    nothing about holding it everywhere.
    ${role_id}=    Grant Scoped Role    Perm Scope Narrow    org    org_id    ${ORG_A}
    Permission Should Be Denied    scope_user    read:iam_user:domain
    [Teardown]    Unassign Role From User    ${SCOPE_USER_ID}    ${role_id}

An Expired Grant Stops Answering Without Any Rebuild
    [Documentation]    Expiry is applied when the permission is READ, not when the cache
    ...    is next rebuilt. An assignment whose expiry has passed must stop answering on
    ...    the next request even though nothing has touched the user since.
    Skip    An assignment expiry cannot be set through the API (ManageUserRoleAssignments takes only add/remove), so the expired grant this case needs cannot be created. The read-time expiry filter is covered by unit tests; re-enable when the assignment API carries expires_at.
