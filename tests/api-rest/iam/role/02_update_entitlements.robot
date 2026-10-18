*** Settings ***
Documentation     Update Role with an entitlement delta: PATCH /v1/iam/roles/:id accepts
...               `entitlements: {add: [{resource_id, action_id|null, description?}], remove: [ids],
...               update: [{id, description}]}`. An update changes only the description.
...
...               The role row and the entitlement changes commit in one transaction with one
...               permission rebuild for the role. A grant takes the resource's max_scope, a
...               null action_id grants every action (`*`), and a delta needs
...               manage_entitlements on iam_role on top of update. The generic engine route
...               /v1/iam/iam_role does not accept the delta.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Role Entitlement Fixtures


*** Variables ***
${ENGINE_ROLE_API}      /v1/iam/iam_role
${READ_EXPRESSION}      read:iam_role:tenant


*** Keywords ***
Set Up Role Entitlement Fixtures
    [Documentation]    A target role held by a plain user, a second role for the
    ...    foreign-entitlement check, and the iam_role resource with its read action.
    Create Authorized API Session
    ${resource_id}    ${action_id}=    Resolve Resource And Action    iam_role    read
    Set Suite Variable    ${IAM_ROLE_RESOURCE_ID}    ${resource_id}
    Set Suite Variable    ${READ_ACTION_ID}    ${action_id}
    ${role_id}=    Create Delta Role    Role Delta Target
    Set Suite Variable    ${TARGET_ROLE_ID}    ${role_id}
    ${role_id}=    Create Delta Role    Role Delta Foreign
    Set Suite Variable    ${FOREIGN_ROLE_ID}    ${role_id}
    ${plain_id}    ${plain_email}=    Create Probe User Session    role_delta_plain
    Set Suite Variable    ${PLAIN_USER_ID}    ${plain_id}
    Assign Role To User    ${PLAIN_USER_ID}    ${TARGET_ROLE_ID}

Create Delta Role
    [Arguments]    ${label}
    ${rname}=    Unique Display Name    ${label}
    ${resp}=    POST On Session    api    ${ROLE_API}
    ...    json=${{ {'name': $rname, 'description': 'Robot role entitlement delta', 'is_requestable': False} }}
    ${role_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    roles    ${role_id}
    RETURN    ${role_id}

Role Etag
    [Arguments]    ${role_id}
    ${resp}=    GET On Session    api    ${ROLE_API}/${role_id}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item][etag]

Role Description
    [Arguments]    ${role_id}
    ${resp}=    GET On Session    api    ${ROLE_API}/${role_id}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item][description]

Patch Role Entitlements
    [Documentation]    Sends an entitlement delta with the role's current etag (or the one
    ...    given) plus any extra role fields. Returns the raw response.
    [Arguments]    ${role_id}    ${add}=${{ [] }}    ${remove}=${{ [] }}    ${alias}=api
    ...    ${etag}=${None}    ${extra}=${{ {} }}    ${update}=${{ [] }}
    IF    $etag is None
        ${etag}=    Role Etag    ${role_id}
    END
    ${delta}=    Evaluate    {'add': $add, 'remove': $remove, 'update': $update}
    ${body}=    Evaluate    {**$extra, 'etag': $etag, 'entitlements': $delta}
    ${resp}=    PATCH On Session    ${alias}    ${ROLE_API}/${role_id}    json=${body}    expected_status=any
    RETURN    ${resp}

Grant
    [Arguments]    ${action_id}=${None}    ${description}=${None}
    ${grant}=    Create Dictionary    resource_id=${IAM_ROLE_RESOURCE_ID}    action_id=${action_id}
    IF    $description is not None
        Set To Dictionary    ${grant}    description=${description}
    END
    RETURN    ${grant}

Role Entitlements
    [Documentation]    Returns {expression: id} for every entitlement the role owns.
    [Arguments]    ${role_id}
    ${graph}=    Set Variable    {"if":["role_id", "=", "${role_id}"]}
    ${resp}=    GET On Session    api    ${ENTITLEMENT_API}
    ...    params=${{ {'graph': $graph, 'fields': 'id,expression', 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${by_expression}=    Evaluate    {e['expression']: e['id'] for e in $resp.json()['items']}
    RETURN    ${by_expression}

Role Should Grant
    [Arguments]    ${role_id}    ${expression}
    ${entitlements}=    Role Entitlements    ${role_id}
    Dictionary Should Contain Key    ${entitlements}    ${expression}
    RETURN    ${entitlements}[${expression}]

Entitlement Description
    [Arguments]    ${ent_id}
    ${graph}=    Set Variable    {"if":["id", "=", "${ent_id}"]}
    ${resp}=    GET On Session    api    ${ENTITLEMENT_API}
    ...    params=${{ {'graph': $graph, 'fields': 'id,description', 'size': 1} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items][0][description]

Role Should Not Grant
    [Arguments]    ${role_id}    ${expression}
    ${entitlements}=    Role Entitlements    ${role_id}
    Dictionary Should Not Contain Key    ${entitlements}    ${expression}

Response Should Be Delta Refusal
    [Arguments]    ${resp}    ${key}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    ${key}


*** Test Cases ***
Entitlement-Only Update Grants An Action And Moves The Etag
    [Documentation]    No role field changes, only a delta: the call still succeeds, bumps
    ...    the etag, and the permission rebuild reaches the role's existing holder.
    [Tags]    entitlements
    Permission Should Be Denied    role_delta_plain    ${READ_EXPRESSION}
    ${etag}=    Role Etag    ${TARGET_ROLE_ID}
    ${grant}=    Grant    ${READ_ACTION_ID}    Robot read grant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}    etag=${etag}
    Response Should Be Update Success    ${resp}    previous_etag=${etag}
    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}
    Permission Should Be Granted    role_delta_plain    ${READ_EXPRESSION}    expected_source_id=${TARGET_ROLE_ID}

A Null Action Grants Every Action
    [Tags]    entitlements
    ${grant}=    Grant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}
    Response Should Be Update Success    ${resp}
    Role Should Grant    ${TARGET_ROLE_ID}    *:iam_role:tenant

Removing An Entitlement Revokes It
    [Tags]    entitlements
    ${ent_id}=    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}
    ${wildcard_id}=    Role Should Grant    ${TARGET_ROLE_ID}    *:iam_role:tenant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    remove=${{ [$ent_id, $wildcard_id] }}
    Response Should Be Update Success    ${resp}
    Role Should Not Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}
    Permission Should Be Denied    role_delta_plain    ${READ_EXPRESSION}

Role Fields And A Delta Change Together
    [Tags]    entitlements
    ${grant}=    Grant    ${READ_ACTION_ID}
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}
    ...    extra=${{ {'description': 'Robot delta with fields'} }}
    Response Should Be Update Success    ${resp}
    ${description}=    Role Description    ${TARGET_ROLE_ID}
    Should Be Equal    ${description}    Robot delta with fields
    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}

A Delta Updates An Entitlement Description
    [Tags]    entitlements
    ${ent_id}=    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}
    ${update}=    Evaluate    [{'id': $ent_id, 'description': 'Robot edited description'}]
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    update=${update}
    Response Should Be Update Success    ${resp}
    ${description}=    Entitlement Description    ${ent_id}
    Should Be Equal    ${description}    Robot edited description
    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}

Updating An Entitlement The Same Delta Removes Is Refused
    [Tags]    entitlements    negative
    ${ent_id}=    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}
    ${update}=    Evaluate    [{'id': $ent_id, 'description': 'Robot must not apply'}]
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    remove=${{ [$ent_id] }}    update=${update}
    Response Should Be Delta Refusal    ${resp}    iam:err_entitlement_update_removed
    Role Should Grant    ${TARGET_ROLE_ID}    ${READ_EXPRESSION}

Removing Another Role's Entitlement Is Refused
    [Tags]    entitlements    negative
    ${grant}=    Grant    ${READ_ACTION_ID}
    ${resp}=    Patch Role Entitlements    ${FOREIGN_ROLE_ID}    add=${{ [$grant] }}
    Response Should Be Update Success    ${resp}
    ${foreign_ent_id}=    Role Should Grant    ${FOREIGN_ROLE_ID}    ${READ_EXPRESSION}
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    remove=${{ [$foreign_ent_id] }}
    Response Should Be Delta Refusal    ${resp}    iam:err_entitlement_not_in_role
    Role Should Grant    ${FOREIGN_ROLE_ID}    ${READ_EXPRESSION}
    ${update}=    Evaluate    [{'id': $foreign_ent_id, 'description': 'Robot must not apply'}]
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    update=${update}
    Response Should Be Delta Refusal    ${resp}    iam:err_entitlement_not_in_role

A Grant The Role Already Has Is Refused And Rolls Back
    [Documentation]    The refusal must also undo the role field change sent with it.
    [Tags]    entitlements    negative
    ${before}=    Role Description    ${TARGET_ROLE_ID}
    ${grant}=    Grant    ${READ_ACTION_ID}
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}
    ...    extra=${{ {'description': 'Robot must roll back'} }}
    Response Should Be Delta Refusal    ${resp}    iam:err_entitlement_already_granted
    ${after}=    Role Description    ${TARGET_ROLE_ID}
    Should Be Equal    ${after}    ${before}

The Same Grant Twice In One Delta Is Refused
    [Tags]    entitlements    negative
    ${grant}=    Grant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant, $grant] }}
    Response Should Be Delta Refusal    ${resp}    iam:err_entitlement_duplicated
    Role Should Not Grant    ${TARGET_ROLE_ID}    *:iam_role:tenant

An Unknown Resource Is A Client Error
    [Tags]    entitlements    negative
    ${grant}=    Create Dictionary    resource_id=${NOT_FOUND_ID}    action_id=${None}
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}
    Response Should Be Field Not Found Error    ${resp}    entitlements.add[0].resource_id

A Stale Etag Refuses The Whole Update
    [Tags]    entitlements    negative
    ${stale}=    Role Etag    ${TARGET_ROLE_ID}
    ${resp}=    PATCH On Session    api    ${ROLE_API}/${TARGET_ROLE_ID}
    ...    json=${{ {'etag': $stale, 'description': 'Robot moves the etag'} }}
    Response Should Be Update Success    ${resp}    previous_etag=${stale}
    ${grant}=    Grant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}    etag=${stale}
    Response Should Be Etag Unmatched Error    ${resp}
    Role Should Not Grant    ${TARGET_ROLE_ID}    *:iam_role:tenant

A Plain User Cannot Change A Role's Entitlements
    [Tags]    entitlements    negative    security
    ${grant}=    Grant
    ${resp}=    Patch Role Entitlements    ${TARGET_ROLE_ID}    add=${{ [$grant] }}    alias=role_delta_plain
    Response Should Be Permission Refusal    ${resp}    Granting through the role update
    Role Should Not Grant    ${TARGET_ROLE_ID}    *:iam_role:tenant

Update Permission Alone Does Not Allow A Delta
    [Documentation]    update on iam_role edits the role's fields; changing what it grants
    ...    also needs manage_entitlements, exactly as the dedicated route asserts.
    [Tags]    entitlements    negative    security
    ${resource_id}    ${update_id}=    Resolve Resource And Action    iam_role    update
    ${editor_role}=    Create Probe Role    Role Delta Editor    ${update_id}    ${resource_id}
    Assign Role To User    ${PLAIN_USER_ID}    ${editor_role}
    ${etag}=    Role Etag    ${FOREIGN_ROLE_ID}
    ${resp}=    PATCH On Session    role_delta_plain    ${ROLE_API}/${FOREIGN_ROLE_ID}
    ...    json=${{ {'etag': $etag, 'description': 'Robot edited by editor'} }}    expected_status=any
    Response Should Be Update Success    ${resp}
    ${grant}=    Grant
    ${resp}=    Patch Role Entitlements    ${FOREIGN_ROLE_ID}    add=${{ [$grant] }}    alias=role_delta_plain
    Response Should Be Permission Refusal    ${resp}    Granting with update permission only
    Role Should Not Grant    ${FOREIGN_ROLE_ID}    *:iam_role:tenant
    [Teardown]    Unassign Role From User    ${PLAIN_USER_ID}    ${editor_role}

The Engine Route Does Not Apply A Delta
    [Documentation]    /v1/iam/iam_role is the generic engine surface; it must not become a
    ...    second way to change what a role grants.
    [Tags]    entitlements    negative    security
    ${etag}=    Role Etag    ${FOREIGN_ROLE_ID}
    ${grant}=    Grant
    ${body}=    Evaluate    {'etag': $etag, 'entitlements': {'add': [$grant], 'remove': []}}
    ${resp}=    PATCH On Session    api    ${ENGINE_ROLE_API}/${FOREIGN_ROLE_ID}    json=${body}
    ...    expected_status=any
    Should Be True    ${resp.status_code} < 500
    ...    msg=The engine PATCH failed with ${resp.status_code}: ${resp.text}
    Role Should Not Grant    ${FOREIGN_ROLE_ID}    *:iam_role:tenant
