*** Settings ***
Documentation     Searching Parties. A party is org-scoped, so every search carries org_id.
...               Graph filters rely on the seeded parties (the "Lead" variants), so they
...               pass on any database.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seeded Parties    50
Test Tags         contacts    party    search


*** Variables ***
${PARTY_SCHEMA}    ${CONTACTS_SCHEMA_DIR}/party.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}    params=${{ {'org_id': $CONTACTS_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${PARTY_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${PARTY_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${PARTY_SCHEMA}    size=50    page=99    item_count=0

Search By Display Name Succeeds
    ${resp}=    GET On Session    api    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': '{"if":["display_name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${PARTY_SCHEMA}    size=50    page=0

Search By Type Succeeds
    [Documentation]    Filtering to companies is how a vendor picker narrows the list, so the
    ...    enum has to be filterable and not merely displayable.
    ${resp}=    GET On Session    api    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': '{"if":["type", "=", "individual"]}'} }}
    Response Should Be Search Success    ${resp}    ${PARTY_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PARTY_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'fields': ['display_name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
