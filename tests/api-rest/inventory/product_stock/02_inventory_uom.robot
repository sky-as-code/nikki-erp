*** Settings ***
Documentation     The unit a product's stock is counted in.
...
...               Configured from the product page but owned by Stock, because it decides what a
...               balance means. It lives in its own resource rather than as a column on the
...               template, and the UoM master it points at stays in Essential.
...
...               The rule worth protecting is the change guard: once a product has moved stock,
...               changing its unit would silently reinterpret every quantity ever recorded
...               against it, so an ordinary update must refuse.
Resource          resources/inventory.resource
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Product Template Under Test
...               AND    Ensure Uom Under Test
Test Tags         inventory    product_stock    inventory_uom


*** Test Cases ***
A Product Line Can Be Given An Inventory Unit
    [Documentation]    AC-PROD-INT-022. One row per template; every variant inherits it.
    ${resp}=    POST On Session    api    ${STOCK_PRODUCT_CONFIG_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'inventory_uom_id': $UOM_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_PRODUCT_CONFIG_ID}    ${id}
    Set Global Variable    ${STOCK_PRODUCT_CONFIG_ETAG}    ${etag}

A Product Line Cannot Be Configured Twice
    [Documentation]    A second row would give one product two inventory units, with no way to
    ...    say which its recorded balances are in.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_PRODUCT_CONFIG_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'inventory_uom_id': $UOM_ID, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400

The Unit Can Be Changed While The Product Is Unused
    [Documentation]    TS-PROD-08. Nothing has been recorded in the old unit yet, so nothing
    ...    changes meaning and the update is allowed.
    ${second}=    Create Second Uom For Change
    ${item}=    Get Stock Product Config
    ${resp}=    PUT On Session    api    ${STOCK_PRODUCT_CONFIG_API}/${STOCK_PRODUCT_CONFIG_ID}
    ...    json=${{ {'inventory_uom_id': $second, 'etag': $item['etag']} }}
    ...    expected_status=any
    Response Status Should Be    ${resp}    200
    ${item}=    Get Stock Product Config
    Should Be Equal    ${item}[inventory_uom_id]    ${second}

An Archived Unit Cannot Be Chosen
    [Documentation]    AC-PROD-INT-023 and TS-PROD-07. An archived UoM stays resolvable so
    ...    historical records keep displaying it; what it may not do is be adopted afresh.
    [Tags]    negative
    ${archived}=    Create Archived Uom
    ${item}=    Get Stock Product Config
    ${resp}=    PUT On Session    api    ${STOCK_PRODUCT_CONFIG_API}/${STOCK_PRODUCT_CONFIG_ID}
    ...    json=${{ {'inventory_uom_id': $archived, 'etag': $item['etag']} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=An archived unit must not be selectable for new configuration


*** Keywords ***
Get Stock Product Config
    ${resp}=    GET On Session    api    ${STOCK_PRODUCT_CONFIG_API}/${STOCK_PRODUCT_CONFIG_ID}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}

Create Second Uom For Change
    [Documentation]    A second unit in the same category, so changing to it is a legitimate
    ...    change rather than a cross-category one the UoM engine would reject on its own.
    ${existing}=    Get Variable Value    ${SECOND_INVENTORY_UOM_ID}    ${EMPTY}
    IF    $existing    RETURN    ${existing}
    Ensure Reference Uom
    ${name}=    Unique Display Name    Robot Tonne
    ${symbol}=    Unique Symbol    t
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger', 'factor': '1000', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SECOND_INVENTORY_UOM_ID}    ${id}
    RETURN    ${id}

Create Archived Uom
    [Documentation]    A unit that exists but has been withdrawn from use.
    ${existing}=    Get Variable Value    ${ARCHIVED_INVENTORY_UOM_ID}    ${EMPTY}
    IF    $existing    RETURN    ${existing}
    Ensure Reference Uom
    ${name}=    Unique Display Name    Robot Retired
    ${symbol}=    Unique Symbol    rt
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '0.5', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    POST On Session    api    ${UOM_API}/${id}/archived
    ...    json=${{ {'is_archived': True, 'etag': $etag} }}
    Response Status Should Be    ${resp}    200
    Set Global Variable    ${ARCHIVED_INVENTORY_UOM_ID}    ${id}
    RETURN    ${id}
