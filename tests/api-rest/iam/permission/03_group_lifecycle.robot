*** Settings ***
Documentation     Grants that reach a user through group membership.
...
...               Same shape as the direct-role suite, one indirection further out: the
...               role is assigned to a group, and the user gets it by being a member.
...               Two events must revoke — losing the membership, and the group losing the
...               role — and each has its own cache path, so both are exercised.
...
...               The last test holds the same permission directly AND through a group at
...               once, which is the case where collapsing grant paths into one cached row
...               used to lose access silently.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Group Lifecycle Fixtures


*** Variables ***
${PROBE_EXPRESSION}     read:iam_group:domain


*** Keywords ***
Set Up Group Lifecycle Fixtures
    Create Authorized API Session
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_group    read
    Set Suite Variable    ${GRP_RESOURCE_ID}    ${resource_id}
    Set Suite Variable    ${GRP_READ_ACTION_ID}    ${action_id}

    ${user_id}    ${email}=    Create Probe User Session    grp_probe_user
    Set Suite Variable    ${GRP_USER_ID}    ${user_id}

    ${gname}=    Unique Display Name    Perm Probe Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'description': {'en-US': 'Robot permission suite group'}, 'owner_id': $user_id} }}
    ${group_id}    ${getag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${group_id}
    Set Suite Variable    ${GRP_GROUP_ID}    ${group_id}

    ${role_id}=    Create Probe Role    Perm Group Reader
    ...    ${GRP_READ_ACTION_ID}    ${GRP_RESOURCE_ID}
    Set Suite Variable    ${GRP_ROLE_ID}    ${role_id}

Add User To Group
    [Arguments]    ${group_id}    ${user_id}
    ${resp}=    POST On Session    api    /v1/iam/groups/${group_id}/manage-users
    ...    json=${{ {'add': [$user_id]} }}
    Response Status Should Be    ${resp}    200

Remove User From Group
    [Arguments]    ${group_id}    ${user_id}
    ${resp}=    POST On Session    api    /v1/iam/groups/${group_id}/manage-users
    ...    json=${{ {'remove': [$user_id]} }}
    Response Status Should Be    ${resp}    200

Assign Role To Group
    [Arguments]    ${group_id}    ${role_id}
    ${resp}=    POST On Session    api    /v1/iam/groups/${group_id}/roles
    ...    json=${{ {'add': [$role_id]} }}
    Response Status Should Be    ${resp}    200

Unassign Role From Group
    [Arguments]    ${group_id}    ${role_id}
    ${resp}=    POST On Session    api    /v1/iam/groups/${group_id}/roles
    ...    json=${{ {'remove': [$role_id]} }}
    Response Status Should Be    ${resp}    200


*** Test Cases ***
Group Membership Alone Grants Nothing
    [Documentation]    Being in a group is not itself a permission. Until the group holds
    ...    a role, a member must be refused.
    Add User To Group    ${GRP_GROUP_ID}    ${GRP_USER_ID}
    Permission Should Be Denied    grp_probe_user    ${PROBE_EXPRESSION}

Assigning A Role To The Group Grants Its Members
    [Documentation]    The grant has to reach existing members, not only people who join
    ...    afterwards — the case where a stale cache would leave a whole group unable to
    ...    do what they were just granted.
    Assign Role To Group    ${GRP_GROUP_ID}    ${GRP_ROLE_ID}
    ${body}=    Permission Should Be Granted    grp_probe_user    ${PROBE_EXPRESSION}    expected_kind=group
    ${resp}=    GET On Session    grp_probe_user    ${GROUP_API}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200
    ...    msg=The probe reported a group grant the API did not honour

The Reported Grant Path Names The Group
    [Documentation]    For an indirect grant the group is the actionable name: it is the
    ...    membership, not the role, that the administrator would remove.
    ${body}=    Probe Permission    grp_probe_user    ${PROBE_EXPRESSION}
    ${via_group}=    Evaluate    [m for m in $body['matches'] if m['source_kind'] == 'group']
    Should Not Be Empty    ${via_group}    msg=A group-derived grant was not reported as "group"
    FOR    ${match}    IN    @{via_group}
        Should Not Be Empty    ${match}[source_name]    msg=The grant path carries no group name
    END

Leaving The Group Revokes Immediately
    [Documentation]    Membership is the grant path here, so losing it must revoke on the
    ...    next request even though the role and the group are untouched.
    Remove User From Group    ${GRP_GROUP_ID}    ${GRP_USER_ID}
    Permission Should Be Denied    grp_probe_user    ${PROBE_EXPRESSION}
    ${resp}=    GET On Session    grp_probe_user    ${GROUP_API}    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Reading groups after leaving the group

Unassigning The Role From The Group Revokes Every Member
    [Documentation]    The other revocation path: the member stays, the role leaves. One
    ...    call must reach every member of the group.
    Add User To Group    ${GRP_GROUP_ID}    ${GRP_USER_ID}
    Permission Should Be Granted    grp_probe_user    ${PROBE_EXPRESSION}

    Unassign Role From Group    ${GRP_GROUP_ID}    ${GRP_ROLE_ID}
    Permission Should Be Denied    grp_probe_user    ${PROBE_EXPRESSION}
    ${resp}=    GET On Session    grp_probe_user    ${GROUP_API}    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Reading groups after the group lost the role

Holding A Permission Directly And Via A Group Reports Both
    [Documentation]    The mixed case. The same permission arrives twice by different
    ...    routes; both must be reported, and dropping either route must leave the other
    ...    still granting. Collapsing them into one cached row is what used to make
    ...    removing one silently revoke the whole permission.
    Assign Role To Group    ${GRP_GROUP_ID}    ${GRP_ROLE_ID}
    ${direct_role_id}=    Create Probe Role    Perm Group Reader Direct
    ...    ${GRP_READ_ACTION_ID}    ${GRP_RESOURCE_ID}
    Assign Role To User    ${GRP_USER_ID}    ${direct_role_id}

    ${body}=    Permission Should Be Granted    grp_probe_user    ${PROBE_EXPRESSION}
    ${kinds}=    Evaluate    sorted({m['source_kind'] for m in $body['matches']})
    Should Be Equal    ${kinds}    ${{ ['direct', 'group'] }}
    ...    msg=Expected both a direct and a group grant path, got ${kinds}

    # Drop the direct one: the group path must still answer.
    Unassign Role From User    ${GRP_USER_ID}    ${direct_role_id}
    ${body}=    Permission Should Be Granted    grp_probe_user    ${PROBE_EXPRESSION}    expected_kind=group
    ${resp}=    GET On Session    grp_probe_user    ${GROUP_API}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Removing the direct role revoked access the group still grants

    # Drop the group one too: now there is nothing left.
    Unassign Role From Group    ${GRP_GROUP_ID}    ${GRP_ROLE_ID}
    Permission Should Be Denied    grp_probe_user    ${PROBE_EXPRESSION}
