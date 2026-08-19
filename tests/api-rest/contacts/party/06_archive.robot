*** Settings ***
Documentation     Archiving the Party under test, rotating the saved etag. The party is
...               unarchived again so the later suites see it live.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Party Under Test
Test Tags         contacts    party    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${PARTY_API}/${PARTY_ID}/archived
    ...    json=${{ {'etag': $PARTY_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PARTY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PARTY_ETAG}    ${etag}

Archived Party Is Still Readable
    [Documentation]    Archiving is visibility, not deletion: a historical purchase order
    ...    still names this party as its vendor, and its detail page must be able to resolve
    ...    the name.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PARTY_API}/${PARTY_ID}/archived
    ...    json=${{ {'etag': $PARTY_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PARTY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PARTY_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PARTY_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $PARTY_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PARTY_API}/${PARTY_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PARTY_API}/${PARTY_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
