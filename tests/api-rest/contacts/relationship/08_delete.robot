*** Settings ***
Documentation     Deleting the Relationship under test — always the LAST suite, doubling as
...               cleanup. A relationship is a leaf: nothing references it, so it deletes
...               cleanly and there is no "delete with referencing records fails" case here.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    Response Should Be Delete Success    ${resp}    count=1

Deleting A Relationship Leaves Both Parties Live
    [Documentation]    The cascade runs from party to relationship, never the other way. A
    ...    link is removable without touching either contact it joined.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    ${resp}=    GET On Session    api    ${PARTY_API}/${TARGET_PARTY_ID}
    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    relationship under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${RELATIONSHIP_ID}    ${EMPTY}
    Set Global Variable    ${RELATIONSHIP_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${RELATIONSHIP_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${RELATIONSHIP_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
