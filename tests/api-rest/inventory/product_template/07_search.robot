*** Settings ***
Documentation     Searching Product Templates — the listing page's own query. Graph filters
...               rely on the seeded templates (the "Lead" variants), so they pass on any
...               database. The "graph" query values are JSON strings that requests
...               URL-encodes via the params dict.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Product Template Under Test
...               AND    Ensure Seeded Product Templates    50
Test Tags         inventory    product_template    search


*** Variables ***
${TEMPLATE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_template.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=99    item_count=0

Search By Name Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search By Category Succeeds
    [Documentation]    Filtering a catalog by category is the listing page's primary facet,
    ...    and it is the query a category detail page issues for its related templates.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["category_id", "=", "' + $PRODUCT_CATEGORY_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search By Status Succeeds
    [Documentation]    BR-PROD-TPL-004: because status is independent of is_archived, a
    ...    discontinued-product listing is a plain status filter rather than an archive one.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["status", "=", "draft"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
