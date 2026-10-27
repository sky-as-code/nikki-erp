*** Settings ***
Documentation     A role managed by the system (managed_by_type = system, e.g. the default
...               "User" role) cannot be changed or removed by a client, owner included: the
...               platform relies on its grants. A client-created role stays editable.
Resource          ../../resources/iam.resource
Test Tags         managed


*** Test Cases ***
System Managed Role Refuses Update
    [Tags]    update    negative
    ${role}=    System Managed Role
    ${resp}=    PATCH On Session    api    ${ROLE_API}/${role}[id]
    ...    json=${{ {'etag': $role['etag'], 'description': 'robot tried to change it'} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_role_system_managed

System Managed Role Refuses Delete
    [Tags]    delete    negative
    ${role}=    System Managed Role
    ${resp}=    DELETE On Session    api    ${ROLE_API}/${role}[id]    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_role_system_managed


*** Keywords ***
System Managed Role
    [Documentation]    The first seeded role managed by the system, resolved at runtime.
    ${resp}=    GET On Session    api    ${ROLE_API}
    ...    params=${{ {'graph': '{"if":["managed_by_type", "=", "system"]}', 'fields': ['id', 'etag', 'name']} }}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[items]    msg=No system-managed role is seeded
    RETURN    ${resp.json()}[items][0]
