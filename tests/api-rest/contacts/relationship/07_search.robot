*** Settings ***
Documentation     Searching Relationships. No org_id parameter: the schema carries no such
...               column (see 01_model_schema.robot), so a search that sent one would be
...               rejected rather than scoped.
...
...               Filtering by party_id is what replaces the old nested route
...               /parties/:party_id/relationships. Note a party can be either end of a link,
...               so "this party's relationships" is two queries — party_id and
...               target_party_id — which is exactly what the two edges on the schema model.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    search


*** Variables ***
${RELATIONSHIP_SCHEMA}    ${CONTACTS_SCHEMA_DIR}/relationship.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}
    Response Should Be Search Success    ${resp}    ${RELATIONSHIP_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}
    ...    params=${{ {'page': 0, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${RELATIONSHIP_SCHEMA}    size=7    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}    params=${{ {'page': 99} }}
    Response Should Be Search Success    ${resp}    ${RELATIONSHIP_SCHEMA}    size=50    page=99    item_count=0

Search By Source Party Succeeds
    [Documentation]    The replacement for GET /parties/:party_id/relationships, source half.
    ${graph}=    Set Variable    {"if":["party_id","=","${PARTY_ID}"]}
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}
    ...    params=${{ {'graph': $graph, 'size': 100} }}
    Response Should Be Search Success    ${resp}    ${RELATIONSHIP_SCHEMA}    size=100    page=0
    Search Results Should Contain Id    ${resp}    ${RELATIONSHIP_ID}

Search By Target Party Succeeds
    [Documentation]    The other half. A party is named by relationships pointing at it as
    ...    well as by ones it owns, and a UI showing only one direction would hide half the
    ...    links.
    ${graph}=    Set Variable    {"if":["target_party_id","=","${TARGET_PARTY_ID}"]}
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}
    ...    params=${{ {'graph': $graph, 'size': 100} }}
    Response Should Be Search Success    ${resp}    ${RELATIONSHIP_SCHEMA}    size=100    page=0
    Search Results Should Contain Id    ${resp}    ${RELATIONSHIP_ID}

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}
    ...    params=${{ {'fields': ['type', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
