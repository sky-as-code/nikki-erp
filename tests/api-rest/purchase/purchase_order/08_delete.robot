*** Settings ***
Documentation     Deleting a purchase order, and the far more important question of when it may
...               not be deleted. BR 24.
...
...               This suite runs LAST because it removes the order under test. It also depends
...               on the lifecycle suite having cancelled it: an order is deletable only from
...               cancelled, so a delete run earlier would be refused and the cleanup would
...               strand the fixture.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    delete


*** Test Cases ***
Delete Is Refused On A Draft Order
    [Documentation]    BR 24. A draft still carries a code that was allocated to it, and the
    ...    requirement asks for one deletion rule rather than a status-by-status one. Cancelling
    ...    a draft is cheap and leaves the trail.
    ${id}    ${etag}=    Create Purchase Order
    ${resp}=    DELETE On Session    api    ${PURCHASE_ORDER_API}/${id}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order.not_deletable
    [Teardown]    Delete Purchase Order Fixture    ${id}

Delete Is Refused On A Confirmed Order
    [Documentation]    BR 24, and the reason the rule exists. Deleting a confirmed order would
    ...    remove the evidence that the business committed to a purchase; cancelling records
    ...    that it did and then stopped.
    ${id}    ${etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${id}
    ${resp}=    DELETE On Session    api    ${PURCHASE_ORDER_API}/${id}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order.not_deletable
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Status Should Be    ${order}    200
    [Teardown]    Delete Purchase Order Fixture    ${id}

Delete Succeeds Once The Order Is Cancelled
    ${id}    ${etag}=    Create Purchase Order
    POST On Session    api    ${PURCHASE_ORDER_API}/${id}/cancel    json=${{ {'reason': 'not needed'} }}
    ${resp}=    DELETE On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Should Be Delete Success    ${resp}

Delete The Order Under Test
    [Documentation]    Cleanup for the suite's shared fixture, cancelled first because that is
    ...    the only status a delete is accepted from.
    Ensure Purchase Order Under Test
    POST On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}/cancel    json=${{ {'reason': 'suite cleanup'} }}    expected_status=any
    ${resp}=    DELETE On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}    expected_status=any
    Response Status Should Be    ${resp}    200
    Set Global Variable    ${PURCHASE_ORDER_ID}    ${EMPTY}
