*** Settings ***
Documentation     The create_payment action against a deployment with no gateway enabled.
...
...               That is the stock configuration and the one CI can have: every gateway
...               defaults to ENABLED: false. What these tests pin is that asking to pay
...               through an unavailable gateway is refused as a CLIENT error — a 400 naming
...               what is wrong — rather than a 500. The distinction matters because a caller
...               receiving a 500 has no way to tell "this deployment cannot take momo" from
...               "the payment may or may not have gone through".
...
...               Nothing here drives real gateway traffic. That would take real money from a
...               real merchant account and does not belong in a test suite.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Payment Invoice Org
...               AND    Ensure Payment Method Under Test
Test Tags         paymentinvoice    order    create_payment


*** Test Cases ***
Create Payment Through A Disabled Gateway Is A Client Error
    ${resp}=    POST On Session    api    ${ORDER_API}/create_payment
    ...    json=${{ {'payment_method_id': $PAYMENT_METHOD_ID, 'amount': '150000', 'source': 'ROBT', 'org_id': $PAYINV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    ...    msg=An unavailable gateway must be reported as a client error, never a 500

Create Payment Without A Payment Method Fails
    [Documentation]    The action declares no ParamSchema — its params mix order fields with
    ...    method-specific input that no single schema describes — so the body is checked in
    ...    the action itself. This pins that the check happens.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/create_payment
    ...    json=${{ {'amount': '150000', 'org_id': $PAYINV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Create Payment With A Malformed Amount Is Refused Not Defaulted
    [Documentation]    An amount that cannot be read must be refused, never defaulted to zero:
    ...    falling back would turn a malformed request into a free order that the gateway
    ...    happily accepts.
    [Tags]    negative
    FOR    ${amount}    IN    ${EMPTY}    abc    not-a-number
        ${resp}=    POST On Session    api    ${ORDER_API}/create_payment
        ...    json=${{ {'payment_method_id': $PAYMENT_METHOD_ID, 'amount': $amount, 'org_id': $PAYINV_ORG_ID} }}
        ...    expected_status=any
        Should Be Equal As Integers    ${resp.status_code}    400
        ...    msg=A malformed amount must be refused rather than treated as zero
    END

Create Payment With No Amount Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/create_payment
    ...    json=${{ {'payment_method_id': $PAYMENT_METHOD_ID, 'org_id': $PAYINV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Create Payment With An Unknown Payment Method Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ORDER_API}/create_payment
    ...    json=${{ {'payment_method_id': $NOT_FOUND_ID, 'amount': '150000', 'org_id': $PAYINV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

No Order Is Left Behind By A Refused Create
    [Documentation]    A create refused before the gateway was reached must write nothing. The
    ...    order count is compared across the refusal rather than searching for a specific id,
    ...    because a refused create never returns one.
    ${resp}=    GET On Session    api    ${ORDER_API}    params=${{ {'org_id': $PAYINV_ORG_ID, 'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${before}=    Set Variable    ${resp.json()}[total]
    POST On Session    api    ${ORDER_API}/create_payment
    ...    json=${{ {'payment_method_id': $PAYMENT_METHOD_ID, 'amount': '150000', 'org_id': $PAYINV_ORG_ID} }}
    ...    expected_status=any
    ${resp}=    GET On Session    api    ${ORDER_API}    params=${{ {'org_id': $PAYINV_ORG_ID, 'size': 1} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[total]    ${before}
    ...    msg=A create refused before the gateway must not leave an order behind
