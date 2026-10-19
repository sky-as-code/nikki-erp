*** Settings ***
Documentation     Only goods are tracked in stock. A service template's variants hold no stock, so a
...               warehouse reservation for one is refused before availability is ever read — and
...               is_combo does not change that: a service combo is refused the same way.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Warehouse Under Test
Test Tags         inventory    stock_reservation    stockable


*** Variables ***
${RESERVE_API}          /v1/inventory/inventory_stock_reservation/reserve_warehouse_stock
${NOT_STOCKABLE_KEY}    inventory:stock_reservation.not_stockable


*** Test Cases ***
Reserving A Service Variant Is Refused
    [Tags]    negative
    ${variant_id}=    Create Variant Of Type    service    ${False}
    ${resp}=    Reserve One Unit    ${variant_id}
    Response Should Carry Violation    ${resp}    ${NOT_STOCKABLE_KEY}

Reserving A Service Combo Variant Is Refused
    [Tags]    negative
    ${variant_id}=    Create Variant Of Type    service    ${True}
    ${resp}=    Reserve One Unit    ${variant_id}
    Response Should Carry Violation    ${resp}    ${NOT_STOCKABLE_KEY}


*** Keywords ***
Create Variant Of Type
    [Documentation]    A template of the given product type and combo flag, and one variant of it.
    [Arguments]    ${product_type}    ${is_combo}
    Ensure Product Category Under Test
    ${name}=    Unique Display Name    Robot ${product_type} Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type': $product_type, 'is_combo': $is_combo, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${template_id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${key}=    Unique Code    comb
    ${variant_id}    ${etag}=    Create Product Variant    ${template_id}    ${key}
    RETURN    ${variant_id}

Reserve One Unit
    [Arguments]    ${variant_id}
    ${source_id}=    Unique Code    src
    ${line}=    Create Dictionary    source_line_id=1    product_variant_id=${variant_id}    quantity=1
    ${lines}=    Create List    ${line}
    ${resp}=    POST On Session    api    ${RESERVE_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'warehouse_id': $WAREHOUSE_ID, 'source_id': $source_id, 'idempotency_key': $source_id, 'lines': $lines} }}
    ...    expected_status=any
    RETURN    ${resp}

Response Should Carry Violation
    [Documentation]    Asserts the refusal names the given key: a refusal for another reason
    ...    (an unknown warehouse, say) would pass a bare status-code check.
    [Arguments]    ${resp}    ${key}
    Response Status Should Be    ${resp}    400
    ${keys}=    Evaluate    [item.get('key') for item in $resp.json()]
    Should Contain    ${keys}    ${key}    msg=Expected a '${key}' violation; the server sent ${keys}
