*** Settings ***
Documentation     Updating Relationships. The success cases run first (they consume and
...               rotate the saved etag); negatives follow.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    update


*** Test Cases ***
Update Type Succeeds
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    json=${{ {'type': 'emergency', 'etag': $RELATIONSHIP_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${RELATIONSHIP_ETAG}
    IF    $etag is not None    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/relationship.json    200
    Should Be Equal    ${item}[type]    emergency
    Set Global Variable    ${RELATIONSHIP_ETAG}    ${item}[etag]

Update Note Succeeds
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    json=${{ {'note': 'Robot updated note', 'etag': $RELATIONSHIP_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${RELATIONSHIP_ETAG}
    IF    $etag is not None    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    json=${{ {'note': 'Stale', 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $RELATIONSHIP_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${RELATIONSHIP_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $RELATIONSHIP_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
