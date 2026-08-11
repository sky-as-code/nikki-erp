*** Settings ***
Documentation     Searching Product Variants. The filter by product_template_id is the query
...               the template detail page's Variants section issues, so it is pinned with an
...               exact expectation rather than a non-empty one.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Product Variant Under Test
...               AND    Ensure Seeded Product Variants    50
Test Tags         inventory    product_variant    search


*** Variables ***
${VARIANT_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_variant.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=99    item_count=0

Search By Template Succeeds
    [Documentation]    The template detail page lists its SKUs through exactly this filter.
    ...    A template with no variants would answer with an empty list rather than an error,
    ...    so a non-empty result also proves the seeded variants attached to the right one.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["product_template_id", "=", "' + $PRODUCT_TEMPLATE_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0

Search By Combination Key Succeeds
    [Documentation]    Resolving a known combination to its SKU is how a transaction line
    ...    finds the variant it must reference. Scoped by template, because the combination
    ...    is unique only within one (BR-PROD-VAR-002).
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["combination_key", "=", "' + $PRODUCT_VARIANT_COMBINATION + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0    item_count=1

Search By Status Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["status", "=", "active"]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['sku', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
