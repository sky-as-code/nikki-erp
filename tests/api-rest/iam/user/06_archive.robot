*** Settings ***
Documentation     Bruno: IAM/User/User - Archive. Archives and unarchives the user under
...               test, rotating the saved etag.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/archived
    ...    json=${{ {'etag': $USER_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${USER_ETAG}
    IF    $etag is not None    Set Global Variable    ${USER_ETAG}    ${etag}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/archived
    ...    json=${{ {'etag': $USER_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${USER_ETAG}
    IF    $etag is not None    Set Global Variable    ${USER_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${USER_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $USER_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived

Archive With Empty Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/archived
    ...    json=${{ {'etag': '', 'is_archived': True} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag
