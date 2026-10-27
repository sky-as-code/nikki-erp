*** Settings ***
Documentation     Built-in CRUD of iam_service. A service has no email, avatar, org unit or type
...               field: the schema omits them and a read never projects them, even when asked.
Resource          ../../resources/iam.resource
Test Tags         iam    service


*** Test Cases ***
Schema Omits The Human Only Fields
    [Tags]    schema
    ${resp}=    GET On Session    api    ${SERVICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    display_name
    FOR    ${name}    IN    email    avatar_file_id    type    org_unit_id    is_owner
        Dictionary Should Not Contain Key    ${fields}    ${name}
    END

Create Service
    [Tags]    create
    ${name}=    Unique Display Name    Robot Crud Service
    ${resp}=    POST On Session    api    ${SERVICE_API}    json=${{ {'display_name': $name, 'status': 'active'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    services    ${id}
    Set Suite Variable    ${SERVICE_ID}    ${id}
    Set Suite Variable    ${SERVICE_ETAG}    ${etag}
    Set Suite Variable    ${SERVICE_NAME}    ${name}

Create Without Display Name Fails
    [Tags]    create    negative
    ${resp}=    POST On Session    api    ${SERVICE_API}    json=${{ {'status': 'active'} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    display_name

Create With A User Only Field Fails
    [Tags]    create    negative
    ${resp}=    POST On Session    api    ${SERVICE_API}
    ...    json=${{ {'display_name': 'x', 'email': 'svc.refused@e2e.nikki.vn'} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    email

Update Service
    [Tags]    update
    ${name}=    Unique Display Name    Robot Crud Service Renamed
    ${resp}=    PATCH On Session    api    ${SERVICE_API}/${SERVICE_ID}
    ...    json=${{ {'etag': $SERVICE_ETAG, 'display_name': $name} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${SERVICE_ETAG}
    Set Suite Variable    ${SERVICE_ETAG}    ${etag}
    Set Suite Variable    ${SERVICE_NAME}    ${name}

Get Service By Id
    [Tags]    get
    ${resp}=    GET On Session    api    ${SERVICE_API}/${SERVICE_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[item]
    Should Be Equal    ${item}[display_name]    ${SERVICE_NAME}
    Dictionary Should Not Contain Key    ${item}    email
    Dictionary Should Not Contain Key    ${item}    type

Get Never Projects A User Only Column
    [Tags]    get
    ${resp}=    GET On Session    api    ${SERVICE_API}/${SERVICE_ID}
    ...    params=${{ {'fields': 'display_name,email'} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[item]    email

Search Finds The Service
    [Tags]    search
    ${resp}=    GET On Session    api    ${SERVICE_API}
    ...    params=${{ {'graph': '{"if":["id", "=", "' + $SERVICE_ID + '"]}'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${SERVICE_ID}

Archive Service
    [Tags]    archive
    ${resp}=    POST On Session    api    ${SERVICE_API}/${SERVICE_ID}/archived
    ...    json=${{ {'etag': $SERVICE_ETAG, 'is_archived': True} }}    expected_status=any
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${SERVICE_ETAG}
    Set Suite Variable    ${SERVICE_ETAG}    ${etag}

Delete Service
    [Tags]    delete
    ${resp}=    DELETE On Session    api    ${SERVICE_API}/${SERVICE_ID}
    Response Should Be Delete Success    ${resp}
    ${resp}=    GET On Session    api    ${SERVICE_API}/${SERVICE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
