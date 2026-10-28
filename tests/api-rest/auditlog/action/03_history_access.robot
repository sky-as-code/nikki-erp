*** Settings ***
Documentation     The history route is authorized by the target record, not by the audit log:
...               it runs the target resource's own GetById as the caller. A caller who may read
...               the record sees its history; one who may not is refused exactly as the record
...               itself would refuse them. Fields the caller may not read come back nulled and
...               listed in masked_fields.
...
...               The plain-user tests need an account holding ONLY the system `User` role
...               (PLAIN_USER_USERNAME / PLAIN_USER_PASSWORD); without one they skip. That role
...               reads UoM domain-wide but holds no Contacts grant, which gives one allowed and
...               one refused target.
Library           Collections
Library           RequestsLibrary
Resource          resources/auditlog.resource
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Suite Teardown    Delete Audit Party
Test Tags         auditlog    action    history


*** Test Cases ***
History Without Target Fails
    [Tags]    negative
    ${resp}=    Get History Response    ${{ {'org_id': $UOM_ORG_ID} }}
    Response Should Be Missing Fields Error    ${resp}    resource_code    resource_id

History With Negative Page Fails
    [Tags]    negative
    ${resp}=    Get History Response
    ...    ${{ {'org_id': $UOM_ORG_ID, 'resource_code': $AUDIT_RESOURCE_CODE, 'resource_id': $AUDITED_CAT_ID, 'page': '-1'} }}
    Response Should Be Invalid Number Range Error    ${resp}    page

History Of An Unknown Resource Fails
    [Tags]    negative
    ${resp}=    Get History Response
    ...    ${{ {'org_id': $UOM_ORG_ID, 'resource_code': 'robot_no_such_resource', 'resource_id': $AUDITED_CAT_ID} }}
    Response Should Be Field Not Found Error    ${resp}    resource_code

History Of An Org Scoped Record Requires The Org
    [Documentation]    The target's GetById is org-scoped, so the history call is too.
    [Tags]    negative
    ${resp}=    Get History Response
    ...    ${{ {'resource_code': $AUDIT_RESOURCE_CODE, 'resource_id': $AUDITED_CAT_ID} }}
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_org_id_required

History Of A Never Existing Record Is Empty
    [Documentation]    A record with no row answers like a deleted one: the caller holds read on
    ...    the resource in the org, so the (empty) history is theirs to see.
    ${body}=    Get History    ${NOT_FOUND_ID}
    Should Be Equal As Integers    ${body}[total]    0

Masked Fields Carry No Values
    [Documentation]    Whatever the caller may not read is listed in masked_fields and nulled in
    ...    changes. An owner masks nothing, so this asserts the invariant on every entry.
    ${body}=    Get History    ${AUDITED_CAT_ID}
    FOR    ${entry}    IN    @{body}[items]
        FOR    ${field}    IN    @{entry}[masked_fields]
            ${change}=    Change Of    ${entry}    ${field}
            Should Be Equal    ${change}[old]    ${None}
            Should Be Equal    ${change}[new]    ${None}
        END
    END

Admin Search Of Entries Succeeds
    [Documentation]    The plain auditlog_action search stays available behind
    ...    read:auditlog_action, which the administrator roles hold through their wildcard.
    ${graph}=    Evaluate    json.dumps({'and': [{'if': ['resource_code', '=', $AUDIT_RESOURCE_CODE]}, {'if': ['resource_id', '=', $AUDITED_CAT_ID]}]})    modules=json
    ${resp}=    GET On Session    api    ${AUDIT_ACTION_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[items]

Plain User Reads The History Of A Record They May Read
    [Tags]    permission
    Create Plain User Session
    ${body}=    Get History    ${AUDITED_CAT_ID}    session=plain_user
    Should Not Be Empty    ${body}[items]

Plain User Is Refused The History Of A Record They May Not Read
    [Tags]    permission    negative
    Create Plain User Session
    Ensure Contacts Org
    Ensure Party Under Test
    ${resp}=    Get History Response
    ...    ${{ {'org_id': $CONTACTS_ORG_ID, 'resource_code': 'contacts_party', 'resource_id': $PARTY_ID} }}
    ...    plain_user
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=History must be refused exactly as reading the party itself is

Plain User Is Refused The Audit Log Itself
    [Documentation]    read:auditlog_action is for administrators only.
    [Tags]    permission    negative
    Create Plain User Session
    ${resp}=    GET On Session    plain_user    ${AUDIT_ACTION_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403


*** Keywords ***
Create Plain User Session
    [Documentation]    Signs in as the plain-role account under its own alias, once. Skips the
    ...    test when no such account is configured: an unprovisioned environment is a gap in
    ...    coverage, not a failure of the rule under test.
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
    ${token}=    Set Variable    ${resp.json()}[data][access_token]
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    X-NIKKI-ACCESSTOKEN=${token}
    Create Client Cert Session    plain_user    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
    Set Suite Variable    ${PLAIN_USER_SESSION_OPEN}    ${True}

Delete Audit Party
    [Documentation]    The refused-target party exists only for this file.
    ${id}=    Get Variable Value    ${PARTY_ID}    ${EMPTY}
    IF    $id
        DELETE On Session    api    ${PARTY_API}/${id}
        ...    params=${{ {'org_id': $CONTACTS_ORG_ID} }}    expected_status=any
        Set Global Variable    ${PARTY_ID}    ${EMPTY}
    END
