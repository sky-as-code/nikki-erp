*** Settings ***
Documentation     Searching Product Variants. The filter by product_template_id is the query
...               the template detail page's Variants section issues, so it is pinned with an
...               exact expectation rather than a non-empty one.
Resource          resources/inventory.resource
Resource          resources/essential.resource
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

Search Selecting Two-Hop Related Field Succeeds
    [Documentation]    product_template_uom_name is a related computed field whose path crosses
    ...    two edges (product_template, then uom). The query builder projects it through the
    ...    joined unit in the same statement, so the listing shows the unit's name without a
    ...    denormalized copy on the template.
    Give Template Under Test A Unit
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id', 'sku', 'product_template_uom_name'], 'graph': '{"if":["product_template_id", "=", "' + $PRODUCT_TEMPLATE_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${PRODUCT_VARIANT_ID}
    ${item}=    Evaluate    [i for i in $resp.json()['items'] if i['id'] == '${PRODUCT_VARIANT_ID}'][0]
    Should Be Equal    ${item}[product_template_uom_name][en-US]    ${UOM_NAME}

Search Filtering By Two-Hop Related Field Succeeds
    [Documentation]    The same field is filterable: the predicate lands on the unit's name
    ...    column through both joins.
    Give Template Under Test A Unit
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id', 'sku'], 'graph': '{"if":["product_template_uom_name", "*", "' + $UOM_NAME + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${PRODUCT_VARIANT_ID}

Search Filtering By Aggregate Computed Field Succeeds
    [Documentation]    on_hand_quantity sums the variant's quants in SQL; filtering on it makes
    ...    the builder share that aggregate between the projection and the WHERE clause.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id', 'sku', 'on_hand_quantity'], 'graph': '{"and":[{"if":["product_template_id", "=", "' + $PRODUCT_TEMPLATE_ID + '"]},{"if":["on_hand_quantity", ">=", 0]}]}'} }}
    Response Should Be Search Success    ${resp}    ${VARIANT_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${PRODUCT_VARIANT_ID}


*** Keywords ***
Give Template Under Test A Unit
    [Documentation]    Attaches the shared unit to the template under test (idempotent), so the
    ...    two-hop related field has a value to resolve.
    ${done}=    Get Variable Value    ${TEMPLATE_HAS_UOM}    ${EMPTY}
    IF    $done    RETURN
    Ensure Uom Under Test
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Set Global Variable    ${UOM_NAME}    ${resp.json()}[item][name][en-US]
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${template_etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'uom_id': $UOM_ID, 'etag': $template_etag} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${template_etag}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}
    Set Global Variable    ${TEMPLATE_HAS_UOM}    yes
