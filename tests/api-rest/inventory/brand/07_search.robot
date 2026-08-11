*** Settings ***
Documentation     Searching Brands. Unlike product type, a brand is org-scoped, so every
...               search carries org_id. Graph filters rely on the seeded brands (the
...               "Lead" variants), so they pass on any database.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seeded Brands    50
Test Tags         inventory    brand    search


*** Variables ***
${BRAND_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/brand.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${BRAND_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${BRAND_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${BRAND_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${BRAND_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${BRAND_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${BRAND_SCHEMA}    size=50    page=99    item_count=0

Search By Name Succeeds
    ${resp}=    GET On Session    api    ${BRAND_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${BRAND_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${BRAND_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
