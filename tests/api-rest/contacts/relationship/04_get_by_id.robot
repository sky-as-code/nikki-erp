*** Settings ***
Documentation     Reading a single Relationship.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/relationship.json    200
    Should Be Equal    ${item}[party_id]    ${PARTY_ID}
    Should Be Equal    ${item}[target_party_id]    ${TARGET_PARTY_ID}
    Set Global Variable    ${RELATIONSHIP_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    params=${{ {'fields': ['party_id', 'target_party_id', 'type']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${RELATIONSHIP_ID}
    ...    params=${{ {'fields': ['type', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
