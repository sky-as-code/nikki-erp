*** Settings ***
Documentation     Deleting the Vendor Profile under test — always the LAST suite, doubling as
...               cleanup.
...
...               Deleting a profile is how a party stops being a vendor at all, as distinct
...               from archiving it, which retires it while keeping the record. Both are
...               permitted here; which one a deployment should use once purchase orders
...               reference vendors is a Purchase question, since Purchase holds vendor_id as a
...               plain ulid with no foreign key and so nothing at the database level prevents
...               the delete.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Deleting A Profile Leaves The Party Live
    [Documentation]    The cascade runs from party to profile, never the other way. A company
    ...    that is no longer a supplier is still a contact.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200

Party Is No Longer A Vendor
    [Documentation]    The inverse of the create-time rule: with the row gone, "is this party a
    ...    vendor?" answers no. That is the property the sidecar design exists to give, so it
    ...    is asserted rather than assumed.
    ${graph}=    Set Variable    {"if":["party_id","=","${PARTY_ID}"]}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': $graph, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[items]
    ...    msg=After deleting the profile the party must no longer resolve as a vendor

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    profile under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${VENDOR_PROFILE_ID}    ${EMPTY}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${VENDOR_PROFILE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${VENDOR_PROFILE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
