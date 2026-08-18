*** Settings ***
Documentation     The adversarial pass: everything here expects a refusal.
...
...               The other suites prove that grants work. This one probes the ways a
...               signed-in user might try to grant themselves something, read somebody
...               else's access, or slip past the hand-written routes by going through the
...               generic engine routes instead.
...
...               Two properties matter most. First, the two API surfaces must agree — a
...               permission enforced on `/v1/iam/roles` but not on `/v1/iam/iam_role` is
...               not enforced at all. Second, the probe must never become an oracle: it
...               answers about the caller and nobody else, and a refusal says only "no".
Resource          ../../resources/iam.resource
Suite Setup       Set Up Escalation Fixtures


*** Keywords ***
Set Up Escalation Fixtures
    [Documentation]    Two ordinary users with their own sessions, plus one role that
    ...    neither of them holds — the thing they will try and fail to award themselves.
    Create Authorized API Session
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_role    read
    Set Suite Variable    ${ESC_RESOURCE_ID}    ${resource_id}
    Set Suite Variable    ${ESC_ACTION_ID}    ${action_id}

    ${attacker_id}    ${attacker_email}=    Create Probe User Session    attacker
    Set Suite Variable    ${ATTACKER_ID}    ${attacker_id}
    Set Suite Variable    ${ATTACKER_EMAIL}    ${attacker_email}

    ${victim_id}    ${victim_email}=    Create Probe User Session    victim
    Set Suite Variable    ${VICTIM_ID}    ${victim_id}

    ${role_id}=    Create Probe Role    Perm Escalation Target    ${ESC_ACTION_ID}    ${ESC_RESOURCE_ID}
    Set Suite Variable    ${ESC_ROLE_ID}    ${role_id}

    # The victim genuinely holds the role; the attacker never does. Any probe or call
    # by the attacker that reflects the victim's access is a leak.
    Assign Role To User    ${VICTIM_ID}    ${ESC_ROLE_ID}

Request Should Be Refused
    [Documentation]    Asserts a refusal rather than one specific code: 401, 403 and 404 are
    ...    all acceptable answers to "you may not", and a 404 is sometimes the better one
    ...    because it does not confirm the record exists. A 2xx is a failure; so is a 5xx,
    ...    which would mean the check crashed rather than declined.
    ...
    ...    400 is deliberately NOT accepted. A refusal is not a malformed request, and
    ...    allowing 400 here would let a validation failure — a payload this suite got
    ...    wrong — pass as proof that the permission was enforced.
    [Arguments]    ${resp}    ${what}
    Should Be True    ${resp.status_code} in (401, 403, 404)
    ...    msg=${what} returned ${resp.status_code}; expected a refusal


*** Test Cases ***
A Plain User Cannot Assign Themselves A Role
    [Documentation]    The most direct escalation there is.
    ${resp}=    POST On Session    attacker    /v1/iam/users/${ATTACKER_ID}/roles
    ...    json=${{ {'add': [$ESC_ROLE_ID]} }}    expected_status=any
    Request Should Be Refused    ${resp}    Self-assigning a role
    Permission Should Be Denied    attacker    read:iam_role:domain

A Plain User Cannot Assign A Role Through The Engine Route
    [Documentation]    The generic engine routes serve the same records as the hand-written
    ...    ones. If one refuses and the other allows, the permission is not enforced — it is
    ...    merely inconvenient to bypass.
    ${resp}=    POST On Session    attacker    ${USER_API}/${ATTACKER_ID}/roles
    ...    json=${{ {'add': [$ESC_ROLE_ID]} }}    expected_status=any
    Request Should Be Refused    ${resp}    Self-assigning a role via the engine route
    Permission Should Be Denied    attacker    read:iam_role:domain

A Plain User Cannot Add Themselves To A Group
    ${gname}=    Unique Display Name    Perm Escalation Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'description': {'en-US': 'Escalation target'}, 'owner_id': $VICTIM_ID} }}
    ${group_id}    ${getag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${group_id}

    ${resp}=    POST On Session    attacker    /v1/iam/groups/${group_id}/manage-users
    ...    json=${{ {'add': [$ATTACKER_ID]} }}    expected_status=any
    Request Should Be Refused    ${resp}    Self-adding to a group

A Plain User Cannot Create Roles Or Entitlements
    [Documentation]    Creating the grant is as good as receiving it, so both doors must be shut.
    ${rname}=    Unique Display Name    Perm Escalation Forged Role
    ${resp}=    POST On Session    attacker    ${ROLE_API}
    ...    json=${{ {'name': $rname, 'description': 'forged', 'is_requestable': False} }}
    ...    expected_status=any
    Request Should Be Refused    ${resp}    Creating a role

    ${resp}=    POST On Session    attacker    ${ENTITLEMENT_API}
    ...    json=${{ {'name': $rname, 'role_id': $ESC_ROLE_ID, 'action_id': $ESC_ACTION_ID, 'resource_id': $ESC_RESOURCE_ID, 'scope': 'domain'} }}
    ...    expected_status=any
    Request Should Be Refused    ${resp}    Creating an entitlement

A Plain User Cannot Add An Entitlement To An Existing Role
    [Documentation]    The subtler version: leave the role alone and widen what it grants.
    ${resp}=    POST On Session    attacker    ${ROLE_API}/${ESC_ROLE_ID}/entitlements
    ...    json=${{ {'add': []} }}    expected_status=any
    Request Should Be Refused    ${resp}    Managing a role's entitlements

A Plain User Cannot Make Themselves The Owner
    [Documentation]    `is_owner` short-circuits every permission check, so it must be
    ...    unreachable from a user-supplied payload on both create and update.
    ${resp}=    GET On Session    api    ${USER_API}/${ATTACKER_ID}
    ${etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    PATCH On Session    attacker    ${USER_API}/${ATTACKER_ID}
    ...    json=${{ {'etag': $etag, 'is_owner': True} }}    expected_status=any
    Request Should Be Refused    ${resp}    Setting is_owner on self

    # Even as the administrator, is_owner must not be settable through create: the field
    # is not part of the writable surface.
    ${email}=    Unique Email    perm.forged.owner
    ${name}=    Unique Display Name    Perm Forged Owner
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email, 'is_owner': True} }}    expected_status=any
    IF    ${resp.status_code} == 201
        Track Permission Fixture    users    ${resp.json()}[id]
        ${check}=    GET On Session    api    ${USER_API}/${resp.json()}[id]
        ${is_owner}=    Evaluate    $check.json()['item'].get('is_owner', False)
        Should Not Be True    ${is_owner}
        ...    msg=is_owner was accepted from the create payload; ownership is forgeable
    END

Another User's Grants Never Appear In The Caller's Answer
    [Documentation]    The victim holds the role; the attacker does not. The probe is
    ...    self-only by construction — there is no parameter naming a subject — so this
    ...    checks that the implementation actually honours the caller's identity rather
    ...    than, say, matching on the expression alone.
    Permission Should Be Granted    victim    read:iam_role:domain
    Permission Should Be Denied    attacker    read:iam_role:domain

A Refusal Reveals Nothing About What Exists
    [Documentation]    A denial for a resource that exists and one for a resource that does
    ...    not must be indistinguishable. Otherwise the probe becomes a way to enumerate
    ...    the system's resources without holding any permission at all.
    ${real}=    Probe Permission    attacker    read:iam_role:domain
    ${fake}=    Probe Permission    attacker    read:no_such_resource_here:domain
    Should Be Equal    ${real}    ${fake}
    ...    msg=A refusal distinguishes a real resource from an invented one

An Archived Role Stops Granting Its Holder
    [Documentation]    Archiving is the administrative way to switch access off. Nothing
    ...    cascades on archive, so this only holds because the mutation rebuilds the cache.
    ${resp}=    GET On Session    api    ${ROLE_API}/${ESC_ROLE_ID}
    ${etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    POST On Session    api    ${ROLE_API}/${ESC_ROLE_ID}/archived
    ...    json=${{ {'etag': $etag, 'is_archived': True} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200

    Permission Should Be Denied    victim    read:iam_role:domain
    ${resp}=    Read Roles As    victim
    Response Should Be Permission Refusal    ${resp}    Reading roles under an archived grant

A Suspended User Cannot Sign In Again
    [Documentation]    Suspension must stop the next sign-in. It deliberately does not
    ...    assert anything about the token already issued: revoking live tokens is a
    ...    separate mechanism, and pretending otherwise here would hide that gap.
    ${resp}=    GET On Session    api    ${USER_API}/${ATTACKER_ID}
    ${item}=    Set Variable    ${resp.json()}[item]
    ${resp}=    PATCH On Session    api    ${USER_API}/${ATTACKER_ID}
    ...    json=${{ {'etag': $item['etag'], 'display_name': $item['display_name'], 'email': $ATTACKER_EMAIL, 'status': 'suspended'} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200

    ${password}=    Issue Temp Password    ${ATTACKER_EMAIL}    expect_success=${False}
    ${resp}=    POST On Session    api    ${SIGNIN_API}/start
    ...    json=${{ {'username': $ATTACKER_EMAIL} }}    expected_status=any
    # 400 rather than 403 here, and deliberately so: a suspended account is refused for a
    # business reason, not for lack of permission. The caller holds no entitlement either
    # way, so answering 403 would claim a permission check happened that never did. The
    # body is asserted instead, to prove the refusal is the status check and not some
    # unrelated payload complaint.
    Should Be True    ${resp.status_code} in (400, 401, 403)
    ...    msg=Signing in as a suspended user returned ${resp.status_code}; expected a refusal
    Should Contain    ${resp.text}    err_account_not_active
    ...    msg=Sign-in was refused for a reason other than the account status: ${resp.text}
