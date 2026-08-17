*** Settings ***
Documentation     Reservation over the wire: what reserve and unreserve do to a real balance.
...
...               The invariant under test throughout is AC-STOCK-005: reserving changes only
...               reserved_quantity. Nothing here may move an on-hand figure — that is validate's
...               job alone, and the two being confusable is the failure this separates.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Seed Stock For Reservation
Test Tags         inventory    stock_flows    reserve


*** Test Cases ***
Reserving Changes Only The Reserved Quantity
    [Documentation]    AC-STOCK-005. The goods have not moved, so on-hand must be untouched.
    ${on_hand_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved_before}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    10
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${on_hand_after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved_after}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${on_hand_after}    ${on_hand_before}
    ...    msg=Reserving must not change any on-hand quantity
    ${expected}=    Evaluate    ${reserved_before} + 10
    Should Be Equal As Numbers    ${reserved_after}    ${expected}
    ...    msg=Reserving must raise reserved_quantity by the allocated amount

    Set Suite Variable    ${RESERVED_TRANSFER_ID}    ${id}
    Set Suite Variable    ${RESERVED_MOVE_ID}    ${move_id}

Reserving Makes The Move Assigned
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}/${RESERVED_MOVE_ID}
    Should Be Equal    ${resp.json()}[data][status]    assigned
    ...    msg=A fully reserved move is assigned (BR §4.2.3.8)

Reserving Creates A Move Line
    [Documentation]    The line records which balance the claim was made against, which is what
    ...    lets exactly this reservation be released later rather than some other row's.
    ${resp}=    GET On Session    api    ${STOCK_MOVE_LINE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${mine}=    Evaluate    [i for i in $resp.json()['data']['items'] if i.get('move_id') == $RESERVED_MOVE_ID]
    Should Not Be Empty    ${mine}    msg=Reserving must record its allocation as a move line

Reserving Twice Does Not Double Allocate
    [Documentation]    Reserve is re-entrant: a move asks only for its outstanding quantity, which
    ...    is zero once it is fully assigned. Without that, a retried request would claim the
    ...    stock a second time and quietly hide it from everyone else.
    ${before}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    POST On Session    api    ${STOCK_TRANSFER_API}/${RESERVED_TRANSFER_ID}/reserve
    ...    json=${{ {} }}    expected_status=any
    ${after}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${before}
    ...    msg=Reserving an already-reserved transfer must be a no-op

Unreserving Gives The Stock Back
    [Documentation]    BR §4.2.3.9. Releasing restores availability and leaves on-hand alone: the
    ...    goods never went anywhere.
    ${on_hand_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved_before}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${RESERVED_TRANSFER_ID}/unreserve
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${on_hand_after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved_after}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${on_hand_after}    ${on_hand_before}
    ...    msg=Unreserving must not change on-hand
    ${expected}=    Evaluate    ${reserved_before} - 10
    Should Be Equal As Numbers    ${reserved_after}    ${expected}
    ...    msg=Unreserving must release exactly what was reserved

Reserved Quantity Never Goes Negative
    [Documentation]    STOCK-INV-002. Unreserving again must not drive the figure below zero,
    ...    which would mean more stock was released than was ever held.
    POST On Session    api    ${STOCK_TRANSFER_API}/${RESERVED_TRANSFER_ID}/unreserve
    ...    json=${{ {} }}    expected_status=any
    ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be True    ${reserved} >= 0
    ...    msg=reserved_quantity must never go negative

Reserving More Than Is Available Is Partial Not Failed
    [Documentation]    Partial allocation is a normal outcome, not an error: the move becomes
    ...    partially available and the transfer simply does not become ready.
    ${available}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${excessive}=    Evaluate    int(${available}) + 1000

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    ${excessive}
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${check}=    GET On Session    api    ${STOCK_MOVE_API}/${move_id}
    Should Be Equal    ${check.json()}[data][status]    partially_available
    ...    msg=A move that could not be fully covered is partially available

    ${transfer}=    GET On Session    api    ${STOCK_TRANSFER_API}/${id}
    Should Not Be Equal    ${transfer.json()}[data][status]    ready
    ...    msg=A partly-allocated transfer must not report itself ready

    [Teardown]    Run Keywords
    ...    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Check Availability Takes No Ownership
    [Documentation]    BR §4.2.3.7, AC-STOCK-033. It answers a question and claims nothing, so the
    ...    stock it reports as available must still be available afterwards. A check that quietly
    ...    reserved would make "is it worth trying" indistinguishable from "take it".
    ${reserved_before}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/check_availability
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${reserved_after}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${reserved_after}    ${reserved_before}
    ...    msg=Checking availability must not reserve anything

    [Teardown]    Run Keywords
    ...    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any


*** Keywords ***
Seed Stock For Reservation
    [Documentation]    Puts enough stock in the source location for the reservation tests to have
    ...    something to claim. It runs a real receipt, because that is the only way a balance can
    ...    come into existence.
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    IF    ${on_hand} >= 100    RETURN
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    200
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
