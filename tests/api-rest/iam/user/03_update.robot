*** Settings ***
Documentation     Bruno: IAM/User/User - Update (Basic + Duplicate email). The success
...               case runs first (it consumes and rotates the saved etag); negatives
...               follow.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    update


*** Test Cases ***
Update Succeeds
    ${email}=    Unique Email    robot.updated
    ${name}=    Unique Display Name    Robot Updated User
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_url': 'https://cdn.e2e.nikki.vn/avatars/updated.jpg', 'display_name': $name, 'email': $email, 'etag': $USER_ETAG, 'status': 'invited'} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${USER_ETAG}
    IF    $etag is not None    Set Global Variable    ${USER_ETAG}    ${etag}
    Set Global Variable    ${USER_EMAIL}    ${email}

Update With Empty Strings Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_url': '', 'display_name': '', 'email': '', 'etag': ''} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $USER_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Invalid Fields Fails
    [Tags]    negative
    ${long_name}=    Evaluate    'too-long' + 'g' * 203
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'display_name': $long_name, 'email': 'invalid@', 'etag': 'fake', 'orgIds': ['']} }}    expected_status=any
    Response Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user_update_invalid_fields_error.json    400

Update With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $USER_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Update With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update To Duplicate Email Fails
    [Documentation]    Creates a throwaway user, then PATCHes the user under test with
    ...    its email. The throwaway create must NOT touch ${USER_ETAG}.
    [Tags]    negative
    ${email}=    Unique Email    robot.duplicate
    ${name}=    Unique Display Name    Robot Duplicate User
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email} }}
    ${dup_id}    ${dup_etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'email': $email, 'etag': $USER_ETAG} }}    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    email
