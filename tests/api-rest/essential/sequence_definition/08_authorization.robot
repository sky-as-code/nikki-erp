*** Settings ***
Documentation     Authorization of the three Sequence resources (AC34, AC41). The service refuses;
...               hiding a button in the UI is not the control. The plain-role cases need an
...               account holding only the system User role (PLAIN_USER_USERNAME/PASSWORD) and
...               skip without one.
...
...               AC34 (another tenant) is enforced by the tenant filter every read carries; an
...               id that does not exist in the caller's tenant answers not-found, never 403, so
...               the owner of a foreign record is not disclosed.
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    sequence_definition    authorization


*** Test Cases ***
A Caller Without A Token Is Refused
    [Tags]    negative
    Create Anonymous API Session    anonymous
    FOR    ${api}    IN    ${SEQ_DEFINITION_API}    ${SEQ_COUNTER_API}    ${SEQ_ISSUANCE_API}
        ${resp}=    GET On Session    anonymous    ${api}    expected_status=any
        Response Should Be Invalid Access Token Error    ${resp}
    END
    ${resp}=    POST On Session    anonymous    ${SEQ_PREVIEW_API}    json=${{ {} }}    expected_status=any
    Response Should Be Invalid Access Token Error    ${resp}

An Id Outside The Tenant Reads As Not Found
    [Tags]    negative
    FOR    ${api}    IN    ${SEQ_DEFINITION_API}    ${SEQ_COUNTER_API}    ${SEQ_ISSUANCE_API}
        ${resp}=    GET On Session    api    ${api}/${NOT_FOUND_ID}    expected_status=any
        Response Should Be Not Found Error    ${resp}
    END

A Plain User Cannot List Definitions, Counters Or Issuances
    [Tags]    negative    permission
    Create Plain User Session
    FOR    ${api}    IN    ${SEQ_DEFINITION_API}    ${SEQ_COUNTER_API}    ${SEQ_ISSUANCE_API}
        ${resp}=    GET On Session    plain_user    ${api}    expected_status=any
        Should Be Equal As Integers    ${resp.status_code}    403
        ...    msg=A user holding only the system User role must not read ${api}
    END

A Plain User Cannot Create A Definition
    [Tags]    negative    permission
    Create Plain User Session
    ${body}=    Sequence Definition Body
    ${resp}=    POST On Session    plain_user    ${SEQ_DEFINITION_API}    json=${body}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403


*** Keywords ***
Create Plain User Session
    [Documentation]    Signs in as the plain-role account under its own alias, once; skips the
    ...    test when no such account is configured.
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD
    END
    ${opened}=    Get Variable Value    ${PLAIN_USER_SESSION_OPEN}    ${False}
    IF    $opened    RETURN
    Create Anonymous API Session    alias=plain_user_signin
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $username} }}
    ${attempt_id}=    Set Variable    ${resp.json()}[attempt_id]
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $attempt_id, 'passwords': {'password': $password}} }}
    Should Be True    ${resp.json()}[done]    msg=Plain-user sign-in flow did not complete (done != true)
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    X-NIKKI-ACCESSTOKEN=${resp.json()}[data][access_token]
    Create Client Cert Session    plain_user    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
    Set Suite Variable    ${PLAIN_USER_SESSION_OPEN}    ${True}
