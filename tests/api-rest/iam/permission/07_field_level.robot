*** Settings ***
Documentation     Field-level grants over the wire: `{action}:{resource}.{field}:{scope}`.
...
...               Grants are additive. A caller holding only `update:iam_user.display_name:tenant`
...               may update that one field and nothing else; a caller holding only
...               `read:iam_group.name:tenant` reads a group with every other field blanked (null,
...               listed in meta.masked_fields) and may not filter on them. Reads are proven on
...               iam_group because the system "User" role grants every user read:iam_user:tenant.
...               Delete has no field-level form and keeps needing the resource grant. A fields=
...               projection never answers 403: an unreadable field or edge (or edge.leaf) is null
...               and listed in masked_fields; an edge's data is checked against its destination.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Field Fixtures


*** Test Cases ***
Field Entitlement Expression Names The Field
    [Tags]    field
    ${resp}=    GET On Session    api    ${ENTITLEMENT_API}
    ...    params=${{ {'graph': '{"if":["role_id", "=", "' + $FIELD_UPDATE_ROLE + '"]}', 'fields': ['expression', 'field_name']} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[items][0]
    Should Be Equal    ${item}[expression]    update:iam_user.display_name:tenant
    Should Be Equal    ${item}[field_name]    display_name

Probe Answers For The Granted Field Only
    [Tags]    field    probe
    Permission Should Be Granted    field_user    update:iam_user.display_name:tenant
    Permission Should Be Denied    field_user    update:iam_user.email:tenant
    Permission Should Be Denied    field_user    update:iam_user:tenant

Field Grant Updates Its Own Field
    [Tags]    field    update
    ${etag}=    Etag Of Target
    ${name}=    Unique Display Name    Field Renamed
    ${resp}=    PATCH On Session    field_user    ${USER_API}/${FIELD_TARGET_ID}
    ...    json=${{ {'etag': $etag, 'display_name': $name} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${resp}=    GET On Session    api    ${USER_API}/${FIELD_TARGET_ID}
    Should Be Equal    ${resp.json()}[item][display_name]    ${name}

Field Grant Refuses Another Field
    [Tags]    field    update    negative
    ${etag}=    Etag Of Target
    ${resp}=    PATCH On Session    field_user    ${USER_API}/${FIELD_TARGET_ID}
    ...    json=${{ {'etag': $etag, 'display_name': 'x', 'email': 'field.refused@e2e.nikki.vn'} }}
    ...    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Updating email under a display_name grant
    Should Contain    ${resp.text}    update:iam_user.email:

Delete Needs The Resource Grant
    [Tags]    field    delete    negative
    ${resp}=    DELETE On Session    field_user    ${USER_API}/${FIELD_TARGET_ID}    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Deleting under field grants only

Read Field Grant Masks The Other Fields
    [Tags]    field    read
    ${resp}=    GET On Session    field_user    ${GROUP_API}/${FIELD_GROUP_ID}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Not Be Empty    ${body}[item][name]
    Should Be Equal    ${body}[item][description]    ${None}
    Should Contain    ${body}[meta][masked_fields]    description

Filtering On A Masked Field Is Refused
    [Tags]    field    read    negative
    ${resp}=    GET On Session    field_user    ${GROUP_API}
    ...    params=${{ {'graph': '{"if":["owner_id", "=", "' + $FIELD_USER_ID + '"]}'} }}    expected_status=any
    Response Should Be Permission Refusal    ${resp}    Filtering on an unreadable field
    Should Contain    ${resp.text}    read:iam_group.owner_id:


Scalar Fields Projection Masks The Unreadable Field
    [Tags]    field    read    projection
    ${resp}=    GET On Session    field_user    ${GROUP_API}/${FIELD_GROUP_ID}
    ...    params=${{ {'fields': 'name,description'} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[item][description]    ${None}
    Should Contain    ${resp.json()}[meta][masked_fields]    description

Edge Projection Masks An Unreadable Destination
    [Documentation]    The probe user reads iam_user (system User role) but not iam_orgunit: the edge
    ...    is dropped from the query, answered as null and listed under the requested name.
    [Tags]    field    read    projection    edge
    ${resp}=    GET On Session    field_user    ${USER_API}/${FIELD_TARGET_ID}
    ...    params=${{ {'fields': 'display_name,org_unit.name'} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Not Be Empty    ${body}[item][display_name]
    Should Be Equal    ${body}[item][org_unit]    ${None}
    Should Contain    ${body}[meta][masked_fields]    org_unit.name

Edge Projection Masks In A Search At Top Level
    [Tags]    field    read    projection    edge
    ${graph}=    Evaluate    json.dumps({'if': ['id', '=', $FIELD_TARGET_ID]})    modules=json
    ${resp}=    GET On Session    field_user    ${USER_API}
    ...    params=${{ {'graph': $graph, 'fields': 'display_name,org_unit'} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Contain    ${body}[masked_fields]    org_unit
    Should Be Equal    ${body}[items][0][org_unit]    ${None}

Edge Projection Is Readable With The Destination Grant
    [Tags]    field    read    projection    edge
    ${resp}=    GET On Session    api    ${USER_API}/${FIELD_TARGET_ID}
    ...    params=${{ {'fields': 'display_name,org_unit.name'} }}
    Response Status Should Be    ${resp}    200
    Should Not Contain    ${resp.json()}[meta][masked_fields]    org_unit.name

Meta Schema Needs Only A Valid Token
    [Documentation]    The probe user holds no grant on iam_orgunit, yet reads its schema: the
    ...    schema is caller-independent metadata. Without a token the route still answers 401.
    [Tags]    schema
    ${resp}=    GET On Session    field_user    /v1/iam/iam_orgunit/meta/schema    expected_status=any
    Response Status Should Be    ${resp}    200
    Create Anonymous API Session    alias=field_anonymous
    ${resp}=    GET On Session    field_anonymous    /v1/iam/iam_orgunit/meta/schema    expected_status=any
    Response Status Should Be    ${resp}    401


*** Keywords ***
Set Up Field Fixtures
    Create Authorized API Session
    ${resource_id}    ${update_id}=    Resolve Resource And Action    iam_user    update
    ${group_resource_id}    ${read_id}=    Resolve Resource And Action    iam_group    read
    ${user_id}    ${email}=    Create Probe User Session    field_user
    Set Suite Variable    ${FIELD_USER_ID}    ${user_id}

    ${update_role}=    Create Probe Role    Perm Field Update    ${update_id}    ${resource_id}
    ...    field_name=display_name
    ${read_role}=    Create Probe Role    Perm Field Read    ${read_id}    ${group_resource_id}
    ...    field_name=name
    Set Suite Variable    ${FIELD_UPDATE_ROLE}    ${update_role}
    Assign Role To User    ${user_id}    ${update_role}
    Assign Role To User    ${user_id}    ${read_role}

    ${email}=    Unique Email    perm.field.target
    ${name}=    Unique Display Name    Perm Field Target
    ${resp}=    POST On Session    api    ${USER_API}    json=${{ {'display_name': $name, 'email': $email} }}
    ${target_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    users    ${target_id}
    Set Suite Variable    ${FIELD_TARGET_ID}    ${target_id}

    ${group_name}=    Unique Display Name    Perm Field Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $group_name}, 'description': {'en-US': 'only the admin reads this'}, 'owner_id': $user_id} }}
    ${group_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${group_id}
    Set Suite Variable    ${FIELD_GROUP_ID}    ${group_id}

Etag Of Target
    ${resp}=    GET On Session    api    ${USER_API}/${FIELD_TARGET_ID}
    RETURN    ${resp.json()}[item][etag]
