*** Settings ***
Documentation     Searching Comm Channels. Filtering by party_id is what replaces the old
...               nested route: the hand-written API listed a party's channels at
...               /parties/:party_id/channels, and the engine answers the same question with
...               a graph filter on the field.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
...               AND    Ensure Seeded Comm Channels    50
Test Tags         contacts    comm_channel    search


*** Variables ***
${COMM_CHANNEL_SCHEMA}    ${CONTACTS_SCHEMA_DIR}/comm_channel.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${COMM_CHANNEL_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${COMM_CHANNEL_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${COMM_CHANNEL_SCHEMA}    size=50    page=99    item_count=0

Search By Party Succeeds
    [Documentation]    The replacement for GET /parties/:party_id/channels. Every seeded
    ...    channel hangs off the party under test, so the result set is non-empty and every
    ...    row belongs to that party.
    ${graph}=    Set Variable    {"if":["party_id","=","${PARTY_ID}"]}
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': $graph, 'size': 100} }}
    Response Should Be Search Success    ${resp}    ${COMM_CHANNEL_SCHEMA}    size=100    page=0
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    FOR    ${item}    IN    @{items}
        Should Be Equal    ${item}[party_id]    ${PARTY_ID}
    END

Search By Type Succeeds
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': '{"if":["type", "=", "email"]}'} }}
    Response Should Be Search Success    ${resp}    ${COMM_CHANNEL_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'fields': ['value', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
