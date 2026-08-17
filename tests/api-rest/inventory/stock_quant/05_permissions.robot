*** Settings ***
Documentation     Permission regression for the Inventory Stock entitlement decision.
...
...               The seeds in 0005006_inventory_stock_seeds.sql deliberately insert NO
...               iam_entitlements row for the system `User` role, matching the Products
...               decision and inverting the Essential UoM one. Stock levels are commercially
...               sensitive — what a company holds and where — so visibility follows explicit
...               role assignment. Adding a blanket entitlement later would silently expose
...               every balance, and nothing else in the test tree would notice.
...
...               It needs an account holding ONLY the system `User` role. Provide it through
...               PLAIN_USER_USERNAME / PLAIN_USER_PASSWORD in the environment file; without
...               those the suite skips rather than fails, since no environment can be
...               assumed to have provisioned such a user.
Library           Collections
Library           RequestsLibrary
Resource          resources/inventory.resource
Suite Setup       Create Plain User Session
Test Tags         inventory    stock_quant    permission


*** Test Cases ***
Plain User Is Refused Stock Balance Read
    ${resp}=    GET On Session    plain_user    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read stock balances

Plain User Is Refused Inventory Location Read
    ${resp}=    GET On Session    plain_user    ${INVENTORY_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read stock locations

Plain User Is Refused Stock Operation Type Read
    ${resp}=    GET On Session    plain_user    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read stock operation types

Plain User Is Refused Inventory Location Create
    [Documentation]    Write is refused for the same reason as read: the role carries no
    ...    Inventory entitlement at all, so no action on the resource is permitted.
    ${name}=    Unique Display Name    Robot Forbidden Location
    ${code}=    Unique Code    forbidloc
    ${resp}=    POST On Session    plain_user    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=Permission must be refused before the payload is ever validated


*** Keywords ***
Create Plain User Session
    [Documentation]    Signs in as the plain-role account and opens a session under its own
    ...    alias, leaving the shared "api" session untouched. Skips the whole suite when no
    ...    such account is configured — an unprovisioned environment is a gap in coverage,
    ...    not a failure of the rule under test.
    ...    Duplicated from inventory/products/03_permissions.robot: a suite-local keyword in
    ...    each is how that file already does it, and the two suites guard the same decision
    ...    for different resources.
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD to run the Inventory Stock permission regression
    END
    # The org id is resolved through the privileged session, since the plain user may not
    # be permitted to list organizations either.
    Create Authorized API Session
    Ensure Inventory Org

    Create Anonymous API Session    alias=plain_user_signin
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $username} }}
    ${attempt_id}=    Set Variable    ${resp.json()}[attempt_id]
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $attempt_id, 'passwords': {'password': $password}} }}
    Should Be True    ${resp.json()}[done]    msg=Plain-user sign-in flow did not complete (done != true)
    ${token}=    Set Variable    ${resp.json()}[data][access_token]

    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    Authorization=Bearer ${token}
    Create Client Cert Session    plain_user    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
