*** Settings ***
Documentation     Archiving the Vendor Profile under test, rotating the saved etag. The profile
...               is unarchived again so the later suites see it live.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}/archived
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${VENDOR_PROFILE_ETAG}
    IF    $etag is not None    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}

Archived Profile Is Still Readable
    [Documentation]    Orders already placed against a retired supplier still name it, so the
    ...    profile they resolve must not disappear. Archiving withdraws it from the working set
    ...    without rewriting who was ordered from.
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/vendor_profile.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Archiving A Profile Leaves The Party Live
    [Documentation]    Ceasing to buy from a company does not end the relationship with it —
    ...    it may still be a customer. This is the practical payoff of the sidecar design: the
    ...    vendor role is retired without touching the contact.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[is_archived]    ${False}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}/archived
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${VENDOR_PROFILE_ETAG}
    IF    $etag is not None    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
