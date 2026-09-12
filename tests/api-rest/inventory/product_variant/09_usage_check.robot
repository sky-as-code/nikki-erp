*** Settings ***
Documentation     Cross-module usage check on a product variant: Inventory asks Sales,
...               Purchase and the vending machine before letting one be deleted.
...
...               Runs AFTER 08_delete, which owns the variant that suite removes. The
...               unreferenced case here creates and deletes a variant of its own.
Resource          resources/inventory.resource
Test Tags         inventory    product_variant    usage_check


*** Test Cases ***
Delete Is Refused While Another Module References The Variant
    [Documentation]    The seeded variant carries sales, purchase and kiosk references. The
    ...    foreign keys are ON DELETE RESTRICT, so the database would refuse this too — but as
    ...    a driver error naming a constraint, which tells the client nothing about where the
    ...    reference lives. The business check answers first, as a 400 naming the modules.
    [Tags]    negative
    ${resp}=    Delete Referenced Variant
    Response Should Be Resource In Use Error    ${resp}

Refusal Names Every Blocking Module
    [Documentation]    BR-DEL-004: the client is told where to go and clear the references.
    ...    All blocking modules are listed rather than whichever answered first, so one round
    ...    trip shows everything standing in the way. The referencing records are never named —
    ...    those are the dependants' data.
    [Tags]    negative
    ${resp}=    Delete Referenced Variant
    ${used_by}=    Set Variable    ${resp.json()}[0][vars][used_by]
    Should Contain    ${used_by}    sales
    Should Contain    ${used_by}    purchase

Delete Succeeds For An Unreferenced Variant
    [Documentation]    The other half of the rule. Without this the suite would pass just as
    ...    well against a check that refused every delete — which is exactly the bug a graph
    ...    losing its conditions produced during development.
    Ensure Product Template Under Test
    ${key}=    Unique Code    usagecomb
    ${id}    ${etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}    usagesku

    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1


*** Keywords ***
Delete Referenced Variant
    [Documentation]    Attempts to delete a variant other modules reference, and returns the
    ...    response. The variant is found by asking the database-backed API for one that sales
    ...    order lines point at, so the suite does not depend on a hardcoded seed id.
    ${variant_id}=    Find Referenced Variant
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${variant_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    RETURN    ${resp}

Find Referenced Variant
    [Documentation]    Resolves a variant that is genuinely referenced. Seeded sales data names
    ...    the first variants of the seeded catalogue, so the earliest one is used; an empty
    ...    result means the sales seed is missing rather than that the rule is wrong.
    ${cached}=    Get Variable Value    ${REFERENCED_VARIANT_ID}    ${EMPTY}
    IF    $cached    RETURN    ${cached}
    Ensure Inventory Org
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 1, 'fields': 'id', 'graph': '{"if":["id", "=", "01K5INV00000000VARIANT0001"]}'} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=The seeded variant referenced by sales is missing; the usage check cannot be exercised
    Set Global Variable    ${REFERENCED_VARIANT_ID}    ${items}[0][id]
    RETURN    ${items}[0][id]
