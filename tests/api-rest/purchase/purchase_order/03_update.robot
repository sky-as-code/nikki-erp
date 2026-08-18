*** Settings ***
Documentation     Updating a purchase order, and the fields an update must not be able to
...               reach.
...
...               The second half of this suite is the important one. Every lifecycle-bearing
...               field is no_update, which means the ONLY way to change an order's status is
...               through the action that owns it. Without that, a role holding plain `update`
...               could confirm an order, approve it, and mark it acknowledged — all three of
...               which are separate permissions precisely because they are separate powers.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    update


*** Test Cases ***
Update The Editable Fields
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': $PURCHASE_ORDER_ETAG, 'vendor_reference': 'VQ-2026-0042', 'priority': 'urgent'} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${PURCHASE_ORDER_ETAG}    ${etag}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Should Be Equal    ${order.json()}[vendor_reference]    VQ-2026-0042
    Should Be Equal    ${order.json()}[priority]            urgent

Update Refuses A Stale Etag
    [Documentation]    Optimistic concurrency: two buyers editing the same order must not both
    ...    win, and the one holding the older copy is the one that loses.
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': '1', 'vendor_reference': 'stale write'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update Cannot Change The Status
    [Documentation]    Confirming is not an edit. An update that could set the status would let
    ...    a role with `update` commit the business to a purchase without ever holding the
    ...    `confirm` permission.
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': $PURCHASE_ORDER_ETAG, 'status': 'purchase_order'} }}    expected_status=any
    Response Should Be Client Error    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Should Be Equal    ${order.json()}[status]    rfq

Update Cannot Change The Totals
    [Documentation]    §10.2 and §55.12: the totals come from the lines and a client value is
    ...    ignored, not trusted. An order whose header a client could write would be an order
    ...    whose total nobody can verify against anything.
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': $PURCHASE_ORDER_ETAG, 'total_amount': '999999.00'} }}    expected_status=any
    Response Should Be Client Error    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Should Not Be Equal As Numbers    ${order.json()}[total_amount]    999999.00

Update Cannot Change The Approval Evidence
    [Documentation]    approved_by and approved_at are the record that spending control was
    ...    applied. A client that could write them could name somebody else as the approver of
    ...    its own order, which is exactly the control the field exists to provide.
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': $PURCHASE_ORDER_ETAG, 'approved_by': $PURCHASE_BUYER_ID} }}    expected_status=any
    Response Should Be Client Error    ${resp}

Update Cannot Set The Lock Flag Directly
    [Documentation]    Locking and unlocking are actions because unlock demands a reason for the
    ...    audit trail. A direct write to is_locked would bypass that, leaving the order reopened
    ...    with nothing recording why.
    Ensure Purchase Order Under Test
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    json=${{ {'etag': $PURCHASE_ORDER_ETAG, 'is_locked': ${True}} }}    expected_status=any
    Response Should Be Client Error    ${resp}
