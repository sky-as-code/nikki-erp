*** Settings ***
Documentation     Role-side assignment: manage_users and manage_groups.
...
...               Both endpoints take {add: [ids], remove: [ids]} and run in one transaction
...               with one permission rebuild for the role. They assert exactly what the
...               principal-side endpoints assert (manage_role_assignments on iam_user /
...               iam_group), so a caller refused there is refused here too.
...
...               Membership is verified through the principal-side read endpoints
...               (GET /v1/iam/iam_user/:id/roles, GET /v1/iam/iam_group/:id/roles), so
...               each assertion also proves both ends see the same junction rows.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Role Assignment Fixtures


*** Variables ***
${PROBE_EXPRESSION}     read:iam_role:tenant


*** Keywords ***
Set Up Role Assignment Fixtures
    [Documentation]    An administrator session, a role with no entitlement for plain
    ...    assignment, three users, one group, and a signed-in plain user for the
    ...    refusal and grant checks.
    Create Authorized API Session
    ${role_id}=    Create Assignment Role    Role Side Assign Target
    Set Suite Variable    ${TARGET_ROLE_ID}    ${role_id}
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    3
        ${user_id}=    Create Assignment User    ${index}
        Append To List    ${ids}    ${user_id}
    END
    Set Suite Variable    @{USER_IDS}    @{ids}
    ${group_id}=    Create Assignment Group
    Set Suite Variable    ${TARGET_GROUP_ID}    ${group_id}
    ${plain_id}    ${plain_email}=    Create Probe User Session    role_side_plain
    Set Suite Variable    ${PLAIN_USER_ID}    ${plain_id}

Create Assignment Role
    [Arguments]    ${label}
    ${rname}=    Unique Display Name    ${label}
    ${resp}=    POST On Session    api    ${ROLE_API}
    ...    json=${{ {'name': $rname, 'description': 'Robot role-side assignment', 'is_requestable': False} }}
    ${role_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    roles    ${role_id}
    RETURN    ${role_id}

Create Assignment User
    [Arguments]    ${index}
    ${email}=    Unique Email    role.side.user${index}
    ${name}=    Unique Display Name    Role Side User ${index}
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email} }}
    ${user_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    users    ${user_id}
    RETURN    ${user_id}

Create Assignment Group
    ${gname}=    Unique Display Name    Role Side Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'description': {'en-US': 'Role-side assignment'}, 'owner_id': $USER_IDS[0]} }}
    ${group_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${group_id}
    RETURN    ${group_id}

Manage Role Users
    [Arguments]    ${role_id}    ${add}=${{ [] }}    ${remove}=${{ [] }}    ${alias}=api
    ${resp}=    POST On Session    ${alias}    ${ROLE_API}/${role_id}/manage_users
    ...    json=${{ {'add': $add, 'remove': $remove} }}    expected_status=any
    RETURN    ${resp}

Manage Role Groups
    [Arguments]    ${role_id}    ${add}=${{ [] }}    ${remove}=${{ [] }}    ${alias}=api
    ${resp}=    POST On Session    ${alias}    ${ROLE_API}/${role_id}/manage_groups
    ...    json=${{ {'add': $add, 'remove': $remove} }}    expected_status=any
    RETURN    ${resp}

User Should Hold Role
    [Arguments]    ${user_id}    ${role_id}
    ${resp}=    GET On Session    api    /v1/iam/iam_user/${user_id}/roles
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${role_id}

User Should Not Hold Role
    [Arguments]    ${user_id}    ${role_id}
    ${resp}=    GET On Session    api    /v1/iam/iam_user/${user_id}/roles    expected_status=any
    IF    ${resp.status_code} == 200
        Search Results Should Not Contain Id    ${resp}    ${role_id}
    ELSE
        Should Be True    ${resp.status_code} in (400, 404)
        ...    msg=Listing the user's roles returned ${resp.status_code}
    END

Group Should Hold Role
    [Arguments]    ${group_id}    ${role_id}
    ${resp}=    GET On Session    api    /v1/iam/iam_group/${group_id}/roles
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${role_id}

Group Should Not Hold Role
    [Arguments]    ${group_id}    ${role_id}
    ${resp}=    GET On Session    api    /v1/iam/iam_group/${group_id}/roles    expected_status=any
    IF    ${resp.status_code} == 200
        Search Results Should Not Contain Id    ${resp}    ${role_id}
    ELSE
        Should Be True    ${resp.status_code} in (400, 404)
        ...    msg=Listing the group's roles returned ${resp.status_code}
    END


*** Test Cases ***
Manage Users Adds One User
    [Tags]    manage_users
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}    add=${{ [$USER_IDS[0]] }}
    Response Should Be Update Success    ${resp}    count=1
    User Should Hold Role    ${USER_IDS}[0]    ${TARGET_ROLE_ID}

Adding An Already Assigned User Is Accepted
    [Documentation]    Same as the user-side endpoint: the duplicate junction row is
    ...    ignored rather than reported, and the assignment stays in place.
    [Tags]    manage_users
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}    add=${{ [$USER_IDS[0]] }}
    Response Status Should Be    ${resp}    200
    User Should Hold Role    ${USER_IDS}[0]    ${TARGET_ROLE_ID}

Manage Users Adds And Removes In Bulk
    [Documentation]    One call adds two users and removes the one added earlier.
    [Tags]    manage_users    bulk
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}
    ...    add=${{ [$USER_IDS[1], $USER_IDS[2]] }}    remove=${{ [$USER_IDS[0]] }}
    Response Should Be Update Success    ${resp}    count=3
    User Should Not Hold Role    ${USER_IDS}[0]    ${TARGET_ROLE_ID}
    User Should Hold Role    ${USER_IDS}[1]    ${TARGET_ROLE_ID}
    User Should Hold Role    ${USER_IDS}[2]    ${TARGET_ROLE_ID}

Manage Users Removes Users
    [Tags]    manage_users    bulk
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}    remove=${{ [$USER_IDS[1], $USER_IDS[2]] }}
    Response Should Be Update Success    ${resp}    count=2
    User Should Not Hold Role    ${USER_IDS}[1]    ${TARGET_ROLE_ID}
    User Should Not Hold Role    ${USER_IDS}[2]    ${TARGET_ROLE_ID}

Manage Groups Adds And Removes A Group
    [Tags]    manage_groups
    ${resp}=    Manage Role Groups    ${TARGET_ROLE_ID}    add=${{ [$TARGET_GROUP_ID] }}
    Response Should Be Update Success    ${resp}    count=1
    Group Should Hold Role    ${TARGET_GROUP_ID}    ${TARGET_ROLE_ID}
    ${resp}=    Manage Role Groups    ${TARGET_ROLE_ID}    remove=${{ [$TARGET_GROUP_ID] }}
    Response Should Be Update Success    ${resp}    count=1
    Group Should Not Hold Role    ${TARGET_GROUP_ID}    ${TARGET_ROLE_ID}

Assigning From The Role Side Grants And Revokes Immediately
    [Documentation]    The role-side call must rebuild the permission cache in the same
    ...    transaction, exactly like the user-side one: the grant reaches the user on
    ...    their next request and the removal takes it away again.
    [Tags]    manage_users    permission
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_role    read
    ${role_id}=    Create Probe Role    Role Side Reader    ${action_id}    ${resource_id}
    Permission Should Be Denied    role_side_plain    ${PROBE_EXPRESSION}
    ${resp}=    Manage Role Users    ${role_id}    add=${{ [$PLAIN_USER_ID] }}
    Response Should Be Update Success    ${resp}    count=1
    Permission Should Be Granted    role_side_plain    ${PROBE_EXPRESSION}    expected_role_id=${role_id}
    ${resp}=    Manage Role Users    ${role_id}    remove=${{ [$PLAIN_USER_ID] }}
    Response Should Be Update Success    ${resp}    count=1
    Permission Should Be Denied    role_side_plain    ${PROBE_EXPRESSION}

Unknown User Id Is A Client Error
    [Tags]    manage_users    negative
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}    add=${{ [$NOT_FOUND_ID] }}
    Response Should Be Value Not Found Error    ${resp}    ${NOT_FOUND_ID}

Unknown Group Id Is A Client Error
    [Tags]    manage_groups    negative
    ${resp}=    Manage Role Groups    ${TARGET_ROLE_ID}    add=${{ [$NOT_FOUND_ID] }}
    Response Should Be Value Not Found Error    ${resp}    ${NOT_FOUND_ID}

Unknown Role Id Is A Client Error
    [Documentation]    The role is the `:id` path parameter of /iam_role/:id/manage_*, so the
    ...    error names `id`.
    [Tags]    manage_users    manage_groups    negative
    ${resp}=    Manage Role Users    ${NOT_FOUND_ID}    add=${{ [$USER_IDS[0]] }}
    Response Should Be Field Not Found Error    ${resp}    id
    ${resp}=    Manage Role Groups    ${NOT_FOUND_ID}    add=${{ [$TARGET_GROUP_ID] }}
    Response Should Be Field Not Found Error    ${resp}    id

The Same Id In Add And Remove Is A Client Error
    [Tags]    manage_users    negative
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}
    ...    add=${{ [$USER_IDS[0]] }}    remove=${{ [$USER_IDS[0]] }}
    Response Should Be Overlapping Fields Error    ${resp}    add    remove

A System Role Rejects Assignment
    [Documentation]    A system role is held by every user with no assignment row, so
    ...    the endpoint refuses it. Skipped when the environment has no system role.
    [Tags]    manage_users    manage_groups    negative
    ${resp}=    GET On Session    api    ${ROLE_API}
    ...    params=${{ {'graph': '{"if":["is_system", "=", true]}'} }}
    ...    expected_status=any
    Skip If    ${resp.status_code} != 200    Cannot search roles by is_system (${resp.status_code})
    ${items}=    Set Variable    ${resp.json()}[items]
    Skip If    not $items    No system role is seeded in this environment
    ${system_role_id}=    Set Variable    ${items}[0][id]
    ${resp}=    Manage Role Users    ${system_role_id}    add=${{ [$USER_IDS[0]] }}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    iam:err_system_role_assignment_not_allowed
    ${resp}=    Manage Role Groups    ${system_role_id}    add=${{ [$TARGET_GROUP_ID] }}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    iam:err_system_role_assignment_not_allowed

A Plain User Cannot Manage Role Users
    [Documentation]    Includes adding themselves: the role-side route must not become a
    ...    way around the refusal the user-side route gives.
    [Tags]    manage_users    negative    security
    ${resp}=    Manage Role Users    ${TARGET_ROLE_ID}    add=${{ [$PLAIN_USER_ID] }}
    ...    alias=role_side_plain
    Response Should Be Permission Refusal    ${resp}    Self-assigning through the role side
    User Should Not Hold Role    ${PLAIN_USER_ID}    ${TARGET_ROLE_ID}

A Plain User Cannot Manage Role Groups
    [Tags]    manage_groups    negative    security
    ${resp}=    Manage Role Groups    ${TARGET_ROLE_ID}    add=${{ [$TARGET_GROUP_ID] }}
    ...    alias=role_side_plain
    Response Should Be Permission Refusal    ${resp}    Assigning a group through the role side
    Group Should Not Hold Role    ${TARGET_GROUP_ID}    ${TARGET_ROLE_ID}
