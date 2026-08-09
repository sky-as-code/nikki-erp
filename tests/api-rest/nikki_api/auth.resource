*** Settings ***
Documentation     Login-once support: the sign-in flow runs at most once per Robot
...               execution and its tokens are cached in global variables. Suites must
...               use Create Authorized API Session and never call the signin endpoints
...               directly (the iam/authenticate suite is the only exception).
Resource          common.resource


*** Keywords ***
Ensure Logged In
    [Documentation]    Performs the 2-step sign-in on first use; no-op afterwards.
    ${token}=    Get Variable Value    ${API_ACCESS_TOKEN}    ${None}
    IF    $token is not None    RETURN
    Create Anonymous API Session    alias=nikki_signin
    ${resp}=    POST On Session    nikki_signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $SIGNIN_USERNAME} }}
    ${attempt_id}=    Set Variable    ${resp.json()}[attempt_id]
    ${resp}=    POST On Session    nikki_signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $attempt_id, 'passwords': {'password': $SIGNIN_PASSWORD}} }}
    Should Be True    ${resp.json()}[done]    msg=Sign-in flow did not complete (done != true)
    Set Global Variable    ${API_ACCESS_TOKEN}    ${resp.json()}[data][access_token]
    Set Global Variable    ${API_REFRESH_TOKEN}    ${resp.json()}[data][refresh_token]

Create Authorized API Session
    [Documentation]    Session with mTLS client certificate and Bearer token (logging in first if needed).
    [Arguments]    ${alias}=api
    Ensure Logged In
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    Authorization=Bearer ${API_ACCESS_TOKEN}
    Create Client Cert Session    ${alias}    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
