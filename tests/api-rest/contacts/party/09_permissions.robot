*** Settings ***
Documentation     Permission regression for the Contacts entitlement decision.
...
...               Contacts asserted no permissions at all before the engine conversion: both
...               of its resource-code constant sets were dead code with no call sites, and no
...               contacts IAM migration existed. The conversion changed that — the engine
...               asserts using the schema name as the resource code — and
...               0004003_contacts_iam.sql seeds resources and actions but deliberately NO
...               iam_entitlements row for the system `User` role.
...
...               That is the same choice Inventory Products made and the inverse of Essential
...               UoM, and it matters more here: a party record carries a supplier's legal
...               name, tax id and address, and a vendor profile hangs off it. A blanket grant
...               would expose the whole contact book to every authenticated account, and
...               nothing else in the test tree would notice.
...
...               Numbered 09 so it runs after 08_delete: it opens its own session and touches
...               none of the fixtures, but the lifecycle order stays the contract.
...
...               It needs an account holding ONLY the system `User` role. Provide it through
...               PLAIN_USER_USERNAME / PLAIN_USER_PASSWORD in the environment file; without
...               those the suite skips rather than fails, since no environment can be assumed
...               to have provisioned such a user.
Library           Collections
Library           RequestsLibrary
Resource          resources/contacts.resource
Suite Setup       Create Plain User Session
Test Tags         contacts    party    permission


*** Test Cases ***
Plain User Is Refused Party Read
    [Documentation]    The contact book is not universally readable. This is the guard on the
    ...    "no entitlements" half of the seed migration.
    ${resp}=    GET On Session    plain_user    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read the contact book

Plain User Is Refused Comm Channel Read
    [Documentation]    A channel carries a private phone number or address, so it is at least
    ...    as sensitive as the party it belongs to.
    ${resp}=    GET On Session    plain_user    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read communication channels

Plain User Is Refused Relationship Read
    ${resp}=    GET On Session    plain_user    ${RELATIONSHIP_API}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not read party relationships

Plain User Is Refused Party Create
    [Documentation]    Write is refused for the same reason as read: the role carries no
    ...    Contacts entitlement at all, so no action on the resource is permitted. The payload
    ...    is deliberately valid — a 403 here proves permission is asserted before validation,
    ...    which is the engine's documented pipeline order.
    ${name}=    Unique Display Name    Robot Forbidden Party
    ${resp}=    POST On Session    plain_user    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=Permission must be refused before the payload is ever validated

Plain User Is Allowed Uom Read
    [Documentation]    The contrast that gives the tests above their meaning. UoM grants the
    ...    system User role domain-wide read on purpose, so a 403 here would mean the account
    ...    is simply broken rather than that Contacts is correctly restricted.
    ${resp}=    GET On Session    plain_user    /v1/essential/essential_uom
    ...    params=${{ {'size': 1} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    403
    ...    msg=The plain user must still read UoM; a 403 here means the fixture account is wrong


*** Keywords ***
Create Plain User Session
    [Documentation]    Signs in as the plain-role account and opens a session under its own
    ...    alias, leaving the shared "api" session untouched. Skips the whole suite when no
    ...    such account is configured — an unprovisioned environment is a gap in coverage, not
    ...    a failure of the rule under test.
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD to run the Contacts permission regression
    END
    # The org id is resolved through the privileged session, since the plain user may not
    # be permitted to list organizations either.
    Create Authorized API Session
    Ensure Contacts Org

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
