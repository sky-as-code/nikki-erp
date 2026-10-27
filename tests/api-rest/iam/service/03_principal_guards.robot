*** Settings ***
Documentation     Where a service may stand. A service may be an org member and may hold roles
...               (assigned through iam_service/:id/manage_roles or iam_role/:id/manage_services, the
...               same junction rows as a user's). It may not be a group member, a group owner, an
...               org-unit member or a role owner; and each role endpoint refuses the other type.
Resource          ../../resources/iam.resource
Suite Setup       Set Up Guard Fixtures
Suite Teardown    Delete Guard Org Unit
Test Tags         iam    service    guard


*** Variables ***
${ORG_API}              /v1/iam/iam_org
${GUARD_UNIT_ID}        ${EMPTY}


*** Test Cases ***
Group Refuses A Service Member
    [Tags]    negative
    ${resp}=    POST On Session    api    ${GROUP_API}/${GUARD_GROUP_ID}/manage_users
    ...    json=${{ {'add': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_user    field=add

Group Accepts A User Member
    ${resp}=    POST On Session    api    ${GROUP_API}/${GUARD_GROUP_ID}/manage_users
    ...    json=${{ {'add': [$GUARD_USER_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200

Group Refuses A Service Owner
    [Tags]    negative
    ${gname}=    Unique Display Name    Robot Service Owned Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'owner_id': $GUARD_SERVICE_ID} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_user    field=owner_id

Org Unit Refuses A Service Member
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORGUNIT_API}/${GUARD_UNIT_ID}/manage_users
    ...    json=${{ {'org_id': $GUARD_ORG_ID, 'add': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_user    field=add

Org Accepts A Service Member
    ${resp}=    POST On Session    api    ${ORG_API}/${GUARD_ORG_ID}/manage_users
    ...    json=${{ {'add': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${resp}=    POST On Session    api    ${ORG_API}/${GUARD_ORG_ID}/manage_users
    ...    json=${{ {'remove': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200

Role Refuses A Service Owner
    [Tags]    negative
    ${rname}=    Unique Display Name    Robot Service Owned Role
    ${resp}=    POST On Session    api    ${ROLE_API}
    ...    json=${{ {'name': $rname, 'is_requestable': False, 'owner_user_id': $GUARD_SERVICE_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_user    field=owner_user_id

Role Manage Services Refuses A User
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ROLE_API}/${GUARD_ROLE_ID}/manage_services
    ...    json=${{ {'add': [$GUARD_USER_ID]} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_service    field=add

Role Manage Users Refuses A Service
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ROLE_API}/${GUARD_ROLE_ID}/manage_users
    ...    json=${{ {'add': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    iam:err_principal_not_user    field=add

Role Assigns And Removes A Service
    ${resp}=    POST On Session    api    ${ROLE_API}/${GUARD_ROLE_ID}/manage_services
    ...    json=${{ {'add': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${resp}=    GET On Session    api    ${SERVICE_API}/${GUARD_SERVICE_ID}/roles
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${GUARD_ROLE_ID}
    ${resp}=    POST On Session    api    ${ROLE_API}/${GUARD_ROLE_ID}/manage_services
    ...    json=${{ {'remove': [$GUARD_SERVICE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${resp}=    GET On Session    api    ${SERVICE_API}/${GUARD_SERVICE_ID}/roles    expected_status=any
    IF    ${resp.status_code} == 200
        Search Results Should Not Contain Id    ${resp}    ${GUARD_ROLE_ID}
    END

Service Manages Its Own Roles
    ${resp}=    POST On Session    api    ${SERVICE_API}/${GUARD_SERVICE_ID}/manage_roles
    ...    json=${{ {'add': [$GUARD_ROLE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${graph}=    Evaluate    json.dumps({'if': ['roles.id', '=', $GUARD_ROLE_ID]})    modules=json
    ${resp}=    GET On Session    api    ${SERVICE_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${GUARD_SERVICE_ID}
    ${resp}=    POST On Session    api    ${SERVICE_API}/${GUARD_SERVICE_ID}/manage_roles
    ...    json=${{ {'remove': [$GUARD_ROLE_ID]} }}    expected_status=any
    Response Status Should Be    ${resp}    200


*** Keywords ***
Set Up Guard Fixtures
    ${service_id}=    Create Service    Robot Guard Service
    ${user_id}=    Create Plain User    Robot Guard User
    Set Suite Variable    ${GUARD_SERVICE_ID}    ${service_id}
    Set Suite Variable    ${GUARD_USER_ID}    ${user_id}

    ${gname}=    Unique Display Name    Robot Guard Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'owner_id': $user_id} }}
    ${group_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    groups    ${group_id}
    Set Suite Variable    ${GUARD_GROUP_ID}    ${group_id}

    ${rname}=    Unique Display Name    Robot Guard Role
    ${resp}=    POST On Session    api    ${ROLE_API}    json=${{ {'name': $rname, 'is_requestable': False} }}
    ${role_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    roles    ${role_id}
    Set Suite Variable    ${GUARD_ROLE_ID}    ${role_id}

    ${resp}=    GET On Session    api    ${ORG_API}    params=${{ {'size': 1} }}
    Set Suite Variable    ${GUARD_ORG_ID}    ${resp.json()}[items][0][id]
    ${uname}=    Unique Display Name    Robot Guard Unit
    ${resp}=    POST On Session    api    ${ORGUNIT_API}    json=${{ {'name': $uname, 'org_id': $GUARD_ORG_ID} }}
    ${unit_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Suite Variable    ${GUARD_UNIT_ID}    ${unit_id}

Delete Guard Org Unit
    IF    $GUARD_UNIT_ID
        DELETE On Session    api    ${ORGUNIT_API}/${GUARD_UNIT_ID}
        ...    params=${{ {'org_id': $GUARD_ORG_ID} }}    expected_status=any
    END
