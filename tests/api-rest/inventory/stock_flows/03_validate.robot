*** Settings ***
Documentation     Validating an internal transfer: the only operation that moves stock between
...               two real locations (BR §4.2.3.10, AC-STOCK-006, AC-STOCK-035).
...
...               The assertion that matters most is that both ends move together. A source
...               decremented without its destination incremented is stock that has ceased to
...               exist, and no report could explain where it went.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Seed Stock For Validation
Test Tags         inventory    stock_flows    validate


*** Test Cases ***
Validating Moves Stock From Source To Destination
    ${source_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${dest_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    30
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${source_after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${dest_after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}
    ${source_expected}=    Evaluate    ${source_before} - 30
    ${dest_expected}=    Evaluate    ${dest_before} + 30
    Should Be Equal As Numbers    ${source_after}    ${source_expected}
    ...    msg=The source balance must fall by exactly what moved
    Should Be Equal As Numbers    ${dest_after}    ${dest_expected}
    ...    msg=The destination balance must rise by exactly what moved

    Set Suite Variable    ${VALIDATED_TRANSFER_ID}    ${id}
    Set Suite Variable    ${VALIDATED_MOVE_ID}    ${move_id}

Validating Conserves The Total
    [Documentation]    An internal transfer moves stock; it does not create or destroy any. The
    ...    sum across both locations is what a stocktake would have to reconcile against.
    ${source}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${dest}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}
    ${total}=    Evaluate    ${source} + ${dest}
    Should Be Equal As Numbers    ${total}    ${SEEDED_TOTAL}
    ...    msg=An internal transfer must conserve the total quantity across both locations

Validating Consumes The Reservation
    [Documentation]    The reserved quantity is spent, not left standing: the goods it was holding
    ...    have gone. Leaving it would keep hiding stock that is no longer there.
    ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${reserved}    0
    ...    msg=Validating must consume the reservation it moved against

Validating Stamps The Move Lines
    [Documentation]    operation_at is what turns a line from a reservation into a recorded
    ...    movement. Without it there is no way to tell the two apart after the fact.
    ${resp}=    GET On Session    api    ${STOCK_MOVE_LINE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${mine}=    Evaluate    [i for i in $resp.json()['data']['items'] if i.get('move_id') == $VALIDATED_MOVE_ID]
    Should Not Be Empty    ${mine}    msg=The validated move must still have its lines
    FOR    ${line}    IN    @{mine}
        Should Not Be Equal    ${line.get('operation_at')}    ${None}
        ...    msg=A validated move line must carry operation_at
    END

Validating An Unreserved Internal Transfer Moves Nothing
    [Documentation]    An internal move draws from a balance, so with no reservation there is
    ...    nothing to take. It must not silently help itself to whatever is on hand: that stock
    ...    may be promised to someone else.
    ${source_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate    json=${{ {} }}    expected_status=any

    ${source_after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${source_after}    ${source_before}
    ...    msg=Validating without a reservation must not move stock

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any


*** Keywords ***
Seed Stock For Validation
    [Documentation]    Ensures the source location holds enough to move, and records the total
    ...    across both locations so the conservation assertion has a baseline.
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    IF    ${on_hand} < 100
        ${transfer_id}    ${move_id}=    Receive Stock Into Location
        ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    200
        POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
        ...    json=${{ {} }}    expected_status=any
    END
    ${source}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${dest}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}
    ${total}=    Evaluate    ${source} + ${dest}
    Set Suite Variable    ${SEEDED_TOTAL}    ${total}
