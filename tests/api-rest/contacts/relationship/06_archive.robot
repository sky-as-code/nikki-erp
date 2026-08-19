*** Settings ***
Documentation     Archiving the Relationship under test, rotating the saved etag. The
...               relationship is unarchived again so the later suites see it live.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}/archived
    ...    json=${{ {'etag': $RELATIONSHIP_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${RELATIONSHIP_ETAG}
    IF    $etag is not None    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}

Archived Relationship Is Still Readable
    [Documentation]    An employment that ended still happened. Archiving withdraws the link
    ...    from current views without rewriting the history of who worked where.
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/relationship.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Archiving A Relationship Leaves Both Parties Live
    [Documentation]    The link is the archivable thing, not either end of it. Ending an
    ...    employment must not retire the employee or the employer.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[is_archived]    ${False}
    ${resp}=    GET On Session    api    ${PARTY_API}/${TARGET_PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[is_archived]    ${False}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}/archived
    ...    json=${{ {'etag': $RELATIONSHIP_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${RELATIONSHIP_ETAG}
    IF    $etag is not None    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $RELATIONSHIP_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
