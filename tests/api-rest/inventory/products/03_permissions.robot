*** Settings ***
Documentation     Permission regression for the Inventory Products entitlement decision.
...
...               The seeds in 0005003_inventory_product_seeds.sql deliberately insert NO
...               iam_entitlements row for the system `User` role. That is the exact inverse
...               of Essential UoM, which grants every user domain-wide read so anyone can
...               pick a unit in a transaction form. Product master data follows explicit
...               role assignment instead, and this suite is the guard on that choice: adding
...               a blanket entitlement later would silently expose the whole catalog, and
...               nothing else in the test tree would notice.
...
...               It needs an account holding ONLY the system `User` role. Provide it through
...               PLAIN_USER_USERNAME / PLAIN_USER_PASSWORD in the environment file; without
...               those the suite skips rather than fails, since no environment can be
...               assumed to have provisioned such a user.
Library           Collections
Library           RequestsLibrary
Resource          resources/inventory.resource
Suite Setup       Create Plain User Session
Test Tags         inventory    products    permission


*** Test Cases ***
Plain User Is Refused Product Template Read
    [Documentation]    The direct inverse of UoM's behaviour, and the reason task F inserts
    ...    no entitlement for the system User role.
    ${resp}=    GET On Session    plain_user    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read the product catalog

Plain User Is Refused Product Variant Read
    ${resp}=    GET On Session    plain_user    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read product variants

Plain User Is Refused Product Template Create
    [Documentation]    Write is refused for the same reason as read: the role carries no
    ...    Inventory entitlement at all, so no action on the resource is permitted.
    ${name}=    Unique Display Name    Robot Forbidden Template
    ${resp}=    POST On Session    plain_user    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $NOT_FOUND_ID, 'category_id': $NOT_FOUND_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=Permission must be refused before the payload is ever validated

Plain User Is Allowed Uom Read
    [Documentation]    The contrast that gives the tests above their meaning. UoM grants the
    ...    system User role domain-wide read on purpose, so a 403 here would mean the account
    ...    is simply broken rather than that Inventory is correctly restricted.
    ${resp}=    GET On Session    plain_user    /v1/essential/essential_uom
    ...    params=${{ {'size': 1} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    403
    ...    msg=The plain user must still read UoM; a 403 here means the fixture account is wrong


*** Keywords ***
Create Plain User Session
    [Documentation]    Signs in as the plain-role account and opens a session under its own
    ...    alias, leaving the shared "api" session untouched. Skips the whole suite when no
    ...    such account is configured — an unprovisioned environment is a gap in coverage,
    ...    not a failure of the rule under test.
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD to run the Inventory permission regression
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
