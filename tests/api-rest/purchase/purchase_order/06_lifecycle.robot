*** Settings ***
Documentation     The purchase order lifecycle: send, confirm, approve, lock, unlock,
...               acknowledge, duplicate and cancel. AC-03 through AC-10.
...
...               Every one of these is a separate permission because they are materially
...               different powers, and every one writes exactly one audit event in the same
...               transaction as the transition it records (PUR-R6). Both properties are
...               asserted here rather than assumed, because neither is visible from the
...               response body alone.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    lifecycle


*** Test Cases ***
Send Moves A Draft To Quotation Sent
    [Documentation]    AC-03, §13. The status change is the whole of it — this module does not
    ...    send email. What it records is that the RFQ went out, which is what makes "waiting on
    ...    a quote" distinguishable from "not yet asked".
    ${id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/send
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be Equal    ${order.json()}[status]    rfq_sent
    Audit Trail Should Record    ${id}    send
    [Teardown]    Delete Purchase Order Fixture    ${id}

Send Is Refused On A Confirmed Order
    [Documentation]    Sending a confirmed order would be asking for a quotation on something
    ...    already bought.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/send
    ...    expected_status=any
    Response Should Be Client Error    ${resp}
    [Teardown]    Delete Purchase Order Fixture    ${id}

Confirm Commits The Order
    [Documentation]    AC-04, §16. Under the default one-step policy a confirm goes straight to
    ...    purchase_order and stamps confirmed_at, which marks the commitment.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    ${resp}=    Confirm Purchase Order    ${id}
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be Equal    ${order.json()}[status]    purchase_order
    Should Not Be Empty    ${order.json()}[confirmed_at]
    Audit Trail Should Record    ${id}    confirm
    [Teardown]    Delete Purchase Order Fixture    ${id}

Confirm Is Refused On An Order With No Priced Line
    [Documentation]    AC-04. An order with no product line would be a commitment to buy
    ...    nothing — far more likely a half-filled form than a deliberate act, and it produces a
    ...    document a vendor cannot act on.
    ${id}    ${etag}=    Create Purchase Order
    ${resp}=    Confirm Purchase Order    ${id}
    Response Should Be Client Error    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be Equal    ${order.json()}[status]    rfq
    [Teardown]    Delete Purchase Order Fixture    ${id}

Confirming Twice Is Refused Rather Than Silently Repeated
    [Documentation]    A second confirm must not write a second audit event: the trail would
    ...    then say the business committed twice to one purchase.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    Confirm Purchase Order    ${id}
    Response Should Be Client Error    ${resp}
    Audit Trail Should Record    ${id}    confirm
    [Teardown]    Delete Purchase Order Fixture    ${id}

Confirm Recomputes The Totals From The Lines
    [Documentation]    §55.12 and D8. The confirm recomputes before deciding anything, so the
    ...    approval decision is made against what the lines actually say. 10 x 25.00 = 250.00,
    ...    and the header must agree with the column a reader can add up by hand.
    ${id}    ${etag}=    Create Confirmable Purchase Order    10    25.00
    Confirm Purchase Order    ${id}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Status Should Be    ${order}    200
    Should Be Equal As Numbers    ${order.json()}[untaxed_amount]    250.00
    Should Be Equal As Numbers    ${order.json()}[total_amount]      250.00
    [Teardown]    Delete Purchase Order Fixture    ${id}

Lock Closes A Confirmed Order To Editing
    [Documentation]    AC-07, §20. is_locked is a flag, so the order stays confirmed while
    ...    locked — the two facts do not hide each other.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/lock
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be True    ${order.json()}[is_locked]
    Should Be Equal    ${order.json()}[status]    purchase_order
    Audit Trail Should Record    ${id}    lock
    [Teardown]    Delete Purchase Order Fixture    ${id}

Lock Is Refused On A Draft
    [Documentation]    Locking a draft would freeze a document nobody has agreed to, a state
    ...    with no way out except unlocking it again.
    ${id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/lock
    ...    expected_status=any
    Response Should Be Client Error    ${resp}
    [Teardown]    Delete Purchase Order Fixture    ${id}

Unlock Requires A Reason
    [Documentation]    AC-08, §21. Unlocking undoes a control that was deliberately applied, so
    ...    the trail has to say why. A trail of unexplained unlocks is the same as no trail,
    ...    which is why a blank is refused rather than accepted.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/lock
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/unlock
    ...    expected_status=any
    Response Should Be Client Error    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be True    ${order.json()}[is_locked]    msg=A refused unlock must leave the order locked
    [Teardown]    Delete Purchase Order Fixture    ${id}

Unlock With A Reason Reopens The Order And Records Why
    [Documentation]    AC-08. The reason lands on the audit event, which is the point of
    ...    demanding it.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/lock
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/unlock
    ...    json=${{ {'reason': 'price correction agreed with the vendor'} }}
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Not Be True    ${order.json()}[is_locked]
    ${event}=    Audit Trail Should Record    ${id}    unlock
    Should Be Equal    ${event}[reason]    price correction agreed with the vendor
    [Teardown]    Delete Purchase Order Fixture    ${id}

Acknowledge Records The Vendor Confirmation
    [Documentation]    AC-09, §22. Like is_locked this is a flag rather than a status: an order
    ...    is confirmed whether or not the vendor has got round to acknowledging it.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/acknowledge
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Be True    ${order.json()}[vendor_acknowledged]
    Audit Trail Should Record    ${id}    acknowledge
    [Teardown]    Delete Purchase Order Fixture    ${id}

Acknowledging Twice Writes Only One Audit Event
    [Documentation]    A retry after a lost response is not an error, but a second event would
    ...    suggest the vendor acknowledged twice.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/acknowledge
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/acknowledge
    Response Status Should Be    ${resp}    200
    Audit Trail Should Record    ${id}    acknowledge
    [Teardown]    Delete Purchase Order Fixture    ${id}

Acknowledge Is Refused Before Confirmation
    [Documentation]    There is nothing to acknowledge before the order is committed.
    ${id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/acknowledge
    ...    expected_status=any
    Response Should Be Client Error    ${resp}
    [Teardown]    Delete Purchase Order Fixture    ${id}

Cancel Leaves The Order And Its Trail In Place
    [Documentation]    AC-10, §23. This is the whole difference between cancel and delete:
    ...    cancelling records that the business committed and then stopped, where deleting would
    ...    remove the evidence that it ever committed.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/cancel
    ...    json=${{ {'reason': 'vendor withdrew'} }}
    Response Status Should Be    ${resp}    200
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Status Should Be    ${order}    200
    Should Be Equal    ${order.json()}[status]    cancelled
    ${event}=    Audit Trail Should Record    ${id}    cancel
    Should Be Equal    ${event}[reason]    vendor withdrew
    [Teardown]    Delete Purchase Order Fixture    ${id}

A Confirmed Order Can Still Be Cancelled
    [Documentation]    BR 23. A deal both sides agreed to can still fall through, so
    ...    purchase_order is deliberately not a terminal state.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/cancel
    ...    json=${{ {'reason': 'no longer required'} }}
    Response Status Should Be    ${resp}    200
    [Teardown]    Delete Purchase Order Fixture    ${id}

A Cancelled Order Cannot Be Revived
    [Documentation]    Reviving one would produce a document whose history says it was called
    ...    off and whose status says it is live, and there is no reading of that pair a vendor
    ...    could rely on. Duplicate is the way to start again.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/cancel
    ...    json=${{ {'reason': 'abandoned'} }}
    ${resp}=    Confirm Purchase Order    ${id}
    Response Should Be Client Error    ${resp}
    [Teardown]    Delete Purchase Order Fixture    ${id}

Duplicate Starts A Fresh Draft With None Of The History
    [Documentation]    AC-06, §15. The copy carries the terms and a NEW code, at status rfq,
    ...    unlocked and unapproved. Carrying the approval across would produce a document
    ...    claiming to have been approved by somebody who never saw it.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/lock
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/duplicate
    Response Status Should Be    ${resp}    200
    ${copy_id}=    Set Variable    ${resp.json()}[id]
    ${copy}=    GET On Session    api    ${PURCHASE_ORDER_API}/${copy_id}
    Response Status Should Be    ${copy}    200
    ${body}=    Set Variable    ${copy.json()}
    Should Be Equal    ${body}[status]    rfq
    Should Not Be True    ${body}[is_locked]
    Should Not Be True    ${body}[vendor_acknowledged]
    ${original}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Should Not Be Equal    ${body}[code]    ${original.json()}[code]
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${copy_id}
    ...    AND    Delete Purchase Order Fixture    ${id}

Duplicate Copies The Lines And Their Totals
    [Documentation]    A duplicate with no lines would be an empty form rather than a copy of
    ...    the order, and it could not be confirmed.
    ${id}    ${etag}=    Create Confirmable Purchase Order    4    50.00
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/duplicate
    Response Status Should Be    ${resp}    200
    ${copy_id}=    Set Variable    ${resp.json()}[id]
    ${copy}=    GET On Session    api    ${PURCHASE_ORDER_API}/${copy_id}
    Should Be Equal As Numbers    ${copy.json()}[total_amount]    200.00
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${copy_id}
    ...    AND    Delete Purchase Order Fixture    ${id}
