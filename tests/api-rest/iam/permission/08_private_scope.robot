*** Settings ***
Documentation     The private scope over the wire: `{action}:{resource}:private` answers only for a
...               record whose created_by is the caller. A caller holding only private grants reads
...               and updates their own groups, and a search returns only those.
...
...               A private entitlement can only exist on a resource whose min_scope reaches
...               private. No seeded resource does today, so the cases skip rather than fail; the
...               evaluation is proven by the composable and requestguard unit tests.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Private Fixtures
Test Tags         private


*** Test Cases ***
Private Search Returns Only The Caller's Records
    ${ids}=    Group Ids Visible To    private_user
    Should Contain    ${ids}    ${OWN_GROUP_ID}
    Should Not Contain    ${ids}    ${OTHER_GROUP_ID}

Private Grant Updates The Caller's Own Record
    ${etag}=    Etag Of Group    ${OWN_GROUP_ID}
    ${resp}=    PATCH On Session    private_user    ${GROUP_API}/${OWN_GROUP_ID}
    ...    json=${{ {'etag': $etag, 'description': 'updated by its creator'} }}    expected_status=any
    Response Status Should Be    ${resp}    200

Private Grant Refuses Someone Else's Record
    [Tags]    negative
    ${etag}=    Etag Of Group    ${OTHER_GROUP_ID}
    ${resp}=    PATCH On Session    private_user    ${GROUP_API}/${OTHER_GROUP_ID}
    ...    json=${{ {'etag': $etag, 'description': 'not mine'} }}    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Updating a group created by someone else


*** Keywords ***
Set Up Private Fixtures
    Create Authorized API Session
    Skip Unless Private Scope Is Grantable    iam_group
    ${resource_id}    ${create_id}=    Resolve Resource And Action    iam_group    create
    ${resource_id}    ${read_id}=    Resolve Resource And Action    iam_group    read
    ${resource_id}    ${update_id}=    Resolve Resource And Action    iam_group    update
    ${user_id}    ${email}=    Create Probe User Session    private_user
    FOR    ${action_id}    ${scope}    IN    ${create_id}    tenant    ${read_id}    private    ${update_id}    private
        ${role_id}=    Create Probe Role    Perm Private    ${action_id}    ${resource_id}    scope=${scope}
        Assign Role To User    ${user_id}    ${role_id}
    END
    ${own}=    Create Group As    private_user    ${user_id}
    ${other}=    Create Group As    api    ${user_id}
    Set Suite Variable    ${OWN_GROUP_ID}    ${own}
    Set Suite Variable    ${OTHER_GROUP_ID}    ${other}

Skip Unless Private Scope Is Grantable
    [Arguments]    ${resource_code}
    ${resp}=    GET On Session    api    /v1/iam/iam_resource
    ...    params=${{ {'graph': '{"and":[{"if":["code", "=", "' + $resource_code + '"]},{"if":["min_scope", "=", "private"]}]}'} }}
    ...    expected_status=any
    IF    ${resp.status_code} != 200 or len($resp.json().get('items', [])) == 0
        Skip    ${resource_code} does not permit private-scoped entitlements in this deployment
    END

Create Group As
    [Arguments]    ${alias}    ${owner_id}
    ${name}=    Unique Display Name    Perm Private Group
    ${resp}=    POST On Session    ${alias}    ${GROUP_API}    json=${{ {'name': $name, 'owner_id': $owner_id} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${id}
    RETURN    ${id}

Group Ids Visible To
    [Arguments]    ${alias}
    ${resp}=    GET On Session    ${alias}    ${GROUP_API}    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    ${ids}=    Evaluate    [item['id'] for item in $resp.json()['items']]
    RETURN    ${ids}

Etag Of Group
    [Arguments]    ${group_id}
    ${resp}=    GET On Session    api    ${GROUP_API}/${group_id}
    RETURN    ${resp.json()}[item][etag]
