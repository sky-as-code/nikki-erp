*** Settings ***
Documentation     Direct role grants, end to end through the public API.
...
...               A fresh user is created, signed in as themself, and then watched as an
...               administrator grants and removes a role. Every step asserts the probe
...               AND a real API call together: if the probe ever disagreed with what the
...               server actually allows, one of the two assertions would fail. That
...               pairing is the whole point — a probe nobody can trust is worse than no
...               probe at all.
...
...               The last test is the regression for the provenance defect: two roles
...               granting the same entitlement used to collapse into one cached row, so
...               removing either one silently revoked access the user still held.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Role Lifecycle Fixtures


*** Variables ***
${PROBE_EXPRESSION}     read:iam_role:domain


*** Keywords ***
Set Up Role Lifecycle Fixtures
    [Documentation]    An administrator session, a fresh user with their own session, and
    ...    a role granting exactly one readable permission.
    Create Authorized API Session
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_role    read
    Set Suite Variable    ${ROLE_RESOURCE_ID}    ${resource_id}
    Set Suite Variable    ${ROLE_READ_ACTION_ID}    ${action_id}
    ${user_id}    ${email}=    Create Probe User Session    probe_user
    Set Suite Variable    ${PROBE_USER_ID}    ${user_id}
    Set Suite Variable    ${PROBE_USER_EMAIL}    ${email}
    ${role_id}=    Create Probe Role    Perm Role Reader
    ...    ${ROLE_READ_ACTION_ID}    ${ROLE_RESOURCE_ID}
    Set Suite Variable    ${PROBE_ROLE_ID}    ${role_id}


*** Test Cases ***
A New User Starts With Nothing
    [Documentation]    Baseline. A user who has been granted no role must be refused by
    ...    both the probe and the endpoint the permission actually guards.
    Permission Should Be Denied    probe_user    ${PROBE_EXPRESSION}
    ${resp}=    Read Roles As    probe_user
    Response Should Be Permission Refusal    ${resp}    A user with no roles could reach ${ROLE_API}

Assigning A Role Grants Immediately
    [Documentation]    The grant must reach the user on their very next request — not
    ...    after a rebuild, a cache expiry or a fresh sign-in.
    Assign Role To User    ${PROBE_USER_ID}    ${PROBE_ROLE_ID}
    ${body}=    Permission Should Be Granted    probe_user    ${PROBE_EXPRESSION}    expected_kind=direct
    ${resp}=    Read Roles As    probe_user
    Should Be Equal As Integers    ${resp.status_code}    200
    ...    msg=The probe reported a grant the API did not honour

The Reported Grant Path Names The Role
    [Documentation]    Provenance is what makes the answer actionable: an administrator
    ...    needs to know WHICH role to edit, not merely that access exists somewhere.
    ${body}=    Probe Permission    probe_user    ${PROBE_EXPRESSION}
    ${direct}=    Evaluate    [m for m in $body['matches'] if m['source_kind'] == 'direct']
    Should Not Be Empty    ${direct}    msg=A directly assigned role was not reported as "direct"
    FOR    ${match}    IN    @{direct}
        Should Not Be Empty    ${match}[source_name]
        ...    msg=The grant path carries no role name
        Should Contain    ${match}[ent_expression]    iam_role
    END

Removing The Role Revokes Immediately
    [Documentation]    Revocation is the direction that must never lag. The user keeps
    ...    their existing token, so this proves the permission is re-read per request
    ...    rather than baked into the session at sign-in.
    Unassign Role From User    ${PROBE_USER_ID}    ${PROBE_ROLE_ID}
    Permission Should Be Denied    probe_user    ${PROBE_EXPRESSION}
    ${resp}=    Read Roles As    probe_user
    Response Should Be Permission Refusal    ${resp}    Access survived the role being removed

Two Roles Granting The Same Thing Survive Losing One
    [Documentation]    The provenance regression. With one cached row per (user,
    ...    entitlement), two roles granting the same permission collapsed into a single
    ...    row owned by one arbitrary assignment; removing THAT assignment deleted the
    ...    row and revoked a permission the user still held through the other role.
    ...
    ...    With one row per grant path, removing either role must leave the other
    ...    standing — and the probe must report exactly the surviving path.
    ${second_role_id}=    Create Probe Role    Perm Role Reader Duplicate
    ...    ${ROLE_READ_ACTION_ID}    ${ROLE_RESOURCE_ID}
    Assign Role To User    ${PROBE_USER_ID}    ${PROBE_ROLE_ID}
    Assign Role To User    ${PROBE_USER_ID}    ${second_role_id}

    ${body}=    Permission Should Be Granted    probe_user    ${PROBE_EXPRESSION}
    ${paths}=    Get Length    ${body}[matches]
    Should Be True    ${paths} >= 2
    ...    msg=Two roles grant this permission but only ${paths} path(s) were reported

    Unassign Role From User    ${PROBE_USER_ID}    ${PROBE_ROLE_ID}

    # Removing one of two granting roles must not revoke the permission entirely.
    ${body}=    Permission Should Be Granted    probe_user    ${PROBE_EXPRESSION}
    ${resp}=    Read Roles As    probe_user
    Should Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Enforcement revoked access that the surviving role still grants

    Unassign Role From User    ${PROBE_USER_ID}    ${second_role_id}
    Permission Should Be Denied    probe_user    ${PROBE_EXPRESSION}

Archiving A Role Revokes Its Holders
    [Documentation]    Archiving is a revocation that nothing cascades: the rebuild
    ...    filters archived roles out rather than the database deleting rows, so without
    ...    an explicit rebuild the role would keep answering for everyone who holds it.
    Assign Role To User    ${PROBE_USER_ID}    ${PROBE_ROLE_ID}
    Permission Should Be Granted    probe_user    ${PROBE_EXPRESSION}

    ${resp}=    GET On Session    api    ${ROLE_API}/${PROBE_ROLE_ID}
    ${etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    POST On Session    api    ${ROLE_API}/${PROBE_ROLE_ID}/archived
    ...    json=${{ {'etag': $etag, 'is_archived': True} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200

    Permission Should Be Denied    probe_user    ${PROBE_EXPRESSION}
    ${resp}=    Read Roles As    probe_user
    Response Should Be Permission Refusal    ${resp}    An archived role still granted access
