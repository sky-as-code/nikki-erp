*** Settings ***
Documentation     Bruno: IAM/User/User - Delete. Deletes the user under test — always
...               the LAST suite, doubling as cleanup.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${USER_API}/${USER_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so
    ...    the user under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${USER_API}/${USER_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${USER_ID}    ${EMPTY}
    Set Global Variable    ${USER_ETAG}    ${EMPTY}
    Set Global Variable    ${USER_EMAIL}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${USER_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${USER_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
