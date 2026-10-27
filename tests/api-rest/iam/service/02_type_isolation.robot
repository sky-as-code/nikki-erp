*** Settings ***
Documentation     Users and services share iam_users, and each resource sees only its own type:
...               a service id is not found on iam_user and a user id is not found on iam_service,
...               for reads and writes alike, and neither search lists the other type.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Isolation Fixtures
Test Tags         iam    service    isolation


*** Test Cases ***
Service Id Is Not Found On Iam User
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}/${ISO_SERVICE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    ${resp}=    PATCH On Session    api    ${USER_API}/${ISO_SERVICE_ID}
    ...    json=${{ {'etag': 'x', 'display_name': 'hijacked'} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    ${resp}=    DELETE On Session    api    ${USER_API}/${ISO_SERVICE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

User Id Is Not Found On Iam Service
    [Tags]    negative
    ${resp}=    GET On Session    api    ${SERVICE_API}/${ISO_USER_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    ${resp}=    DELETE On Session    api    ${SERVICE_API}/${ISO_USER_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

User Search Never Lists A Service
    ${resp}=    GET On Session    api    ${USER_API}    params=${{ {'graph': $ISO_GRAPH} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${ISO_USER_ID}
    Search Results Should Not Contain Id    ${resp}    ${ISO_SERVICE_ID}

Service Search Never Lists A User
    ${resp}=    GET On Session    api    ${SERVICE_API}    params=${{ {'graph': $ISO_GRAPH} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${ISO_SERVICE_ID}
    Search Results Should Not Contain Id    ${resp}    ${ISO_USER_ID}

A Service Created Through Iam User Becomes A User
    [Documentation]    iam_user creates humans only: a client-sent type is overwritten.
    ${email}=    Unique Email    robot.not.service
    ${name}=    Unique Display Name    Robot Not Service
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email, 'type': 'service'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    users    ${id}
    ${resp}=    GET On Session    api    ${USER_API}/${id}
    Should Be Equal    ${resp.json()}[item][type]    user


*** Keywords ***
Set Up Isolation Fixtures
    ${service_id}=    Create Service    Robot Iso Service
    ${user_id}=    Create Plain User    Robot Iso User
    Set Suite Variable    ${ISO_SERVICE_ID}    ${service_id}
    Set Suite Variable    ${ISO_USER_ID}    ${user_id}
    ${graph}=    Evaluate    json.dumps({'if': ['id', 'in', $service_id, $user_id]})    modules=json
    Set Suite Variable    ${ISO_GRAPH}    ${graph}
