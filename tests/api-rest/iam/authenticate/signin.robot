*** Settings ***
Documentation     Sign-in flow tests (Bruno "Authenticate" folder). This suite TESTS the
...               login sequence, so it uses its own anonymous session and suite-local
...               variables — it must never depend on, or write to, the token globals
...               cached by Ensure Logged In.
Resource          nikki_api/common.resource
Resource          nikki_api/assertions.resource
Suite Setup       Create Anonymous API Session    alias=signin
Test Tags         iam    authenticate


*** Test Cases ***
Sign In Start Returns Attempt Id
    ${resp}=    POST On Session    signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $SIGNIN_USERNAME} }}    expected_status=any
    Should Be True    ${{ $resp.status_code in (200, 201) }}
    ...    msg=Sign-in start failed (${resp.status_code}): ${resp.text}
    Should Not Be Empty    ${resp.json()}[attempt_id]
    Set Suite Variable    ${ATTEMPT_ID}    ${resp.json()}[attempt_id]

Sign In Continue Returns Tokens
    ${resp}=    POST On Session    signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $ATTEMPT_ID, 'passwords': {'password': $SIGNIN_PASSWORD}} }}
    Should Be True    ${resp.json()}[done]    msg=Sign-in flow did not complete (done != true)
    Dictionary Should Contain Key    ${resp.json()}[data]    access_token
    Dictionary Should Contain Key    ${resp.json()}[data]    refresh_token
    Set Suite Variable    ${SIGNIN_REFRESH_TOKEN}    ${resp.json()}[data][refresh_token]

Sign In Refresh Rotates Tokens
    [Documentation]    Note: the refresh response body is FLAT, unlike continue's nested "data".
    ${resp}=    POST On Session    signin    ${SIGNIN_API}/refresh
    ...    json=${{ {'refresh_token': $SIGNIN_REFRESH_TOKEN} }}
    Dictionary Should Contain Key    ${resp.json()}    access_token
    Dictionary Should Contain Key    ${resp.json()}    refresh_token
    Should Not Be Empty    ${resp.json()}[access_token]
