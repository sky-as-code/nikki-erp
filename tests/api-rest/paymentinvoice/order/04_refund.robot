*** Settings ***
Documentation     The refund action's guard rails.
...
...               The full matrix — already refunded, canceled, expired, not paid, over the
...               running total — is covered by unit tests, because reaching those states
...               needs an order that a gateway actually collected. What this suite pins is
...               the part that is reachable over HTTP without one: that the action exists,
...               that it validates its body, and that every refusal is a client error rather
...               than a 500.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Payment Invoice Org
Test Tags         paymentinvoice    order    refund


*** Test Cases ***
Refund Without An Order Or Amount Fails
    [Documentation]    Both are required and both must be reported, so a caller is not left
    ...    fixing them one round trip at a time.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/refund
    ...    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Refund Without An Amount Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/refund
    ...    json=${{ {'order_id': 'VDMCMOM0Q8HABCDEFGH'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Refund With A Malformed Amount Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/refund
    ...    json=${{ {'order_id': 'VDMCMOM0Q8HABCDEFGH', 'amount': 'abc'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Refunding An Unknown Order Is A Client Error
    [Documentation]    The caller quoted an order that does not exist. That is their mistake,
    ...    not a server failure, and the distinction is what tells them to check the id.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/refund
    ...    json=${{ {'order_id': 'NOSUCHORDER00000000', 'amount': '50000'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    ...    msg=An unknown order must be reported as a client error, never a 500

Remove Pos Orders On A Terminal With Nothing Queued Succeeds
    [Documentation]    The action replaces the old service's unauthenticated
    ...    DELETE /mpos/pos-orders/:posId; reaching it here proves it is now behind the
    ...    engine's permission check. A terminal with no prompts queued clears zero of them,
    ...    which is a complete outcome rather than an error: the caller asked for the queue to
    ...    be empty and it is.
    ${resp}=    POST On Session    api    ${ORDER_API}/remove_pos_orders/ROBOT_NO_SUCH_TERMINAL
    ...    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200
    Should Be Equal As Integers    ${resp.json()}[affected_count]    0
