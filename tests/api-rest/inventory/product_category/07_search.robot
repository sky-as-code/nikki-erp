*** Settings ***
Documentation     Searching Product Categories. Unlike product_type, a category is org-scoped
...               (org_id required-for-create), so every search here carries org_id. Graph
...               filters rely on the seeded categories (the "Lead" variants), so they pass
...               on any database. The "graph" query values are JSON strings that requests
...               URL-encodes via the params dict.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Seeded Product Categories    50
Test Tags         inventory    product_category    search


*** Variables ***
${PRODUCT_CATEGORY_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_category.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_CATEGORY_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_CATEGORY_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_CATEGORY_SCHEMA}    size=50    page=99    item_count=0

Search By Name Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_CATEGORY_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
