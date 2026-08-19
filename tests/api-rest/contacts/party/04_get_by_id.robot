*** Settings ***
Documentation     Reading a single Party.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Party Under Test
Test Tags         contacts    party    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Set Global Variable    ${PARTY_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    params=${{ {'fields': ['display_name', 'type', 'tax_id']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    params=${{ {'fields': ['display_name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PARTY_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PARTY_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
