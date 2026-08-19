*** Settings ***
Documentation     Reading a single Comm Channel.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
Test Tags         contacts    comm_channel    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/comm_channel.json    200
    Should Be Equal    ${item}[party_id]    ${PARTY_ID}
    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    params=${{ {'fields': ['party_id', 'type', 'value']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    params=${{ {'fields': ['value', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
