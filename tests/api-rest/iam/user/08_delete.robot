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

A Deleted User Cannot Sign In With Their Old Password
    [Documentation]    Working credentials must stop working once the account is deleted.
    ...
    ...    Scope, stated honestly: this proves the *authentication outcome*, not the row
    ...    cleanup. Sign-in refuses a deleted user at the account lookup, before any
    ...    credential is consulted, so this case passes whether or not the password row
    ...    survived the delete — it was verified to pass with the cascade trigger disabled.
    ...    The cleanup itself is enforced by the database (an AFTER DELETE trigger on
    ...    iam_users, migration 1002006) and proven at the SQL level, because
    ...    `iam_password_stores` is not readable through any API and ids are assigned
    ...    server-side, so no black-box request can observe the difference.
    ...
    ...    It earns its place regardless: it is the regression that would catch a future
    ...    sign-in path that consults credentials before checking the account exists.
    # Provisioning goes through the shared keyword rather than a local POST: a user is
    # created in "draft" and only an active account may hold a password, so activation is
    # part of it. Signing in successfully is the precondition, not the assertion.
    ${probe_id}    ${email}=    Create Probe User Session    creds_probe

    ${resp}=    DELETE On Session    api    ${USER_API}/${probe_id}
    Response Should Be Delete Success    ${resp}    count=1

    Create Anonymous API Session    alias=creds_probe_after
    ${resp}=    POST On Session    creds_probe_after    ${SIGNIN_API}/start
    ...    json=${{ {'username': $email} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A deleted user could still begin sign-in
