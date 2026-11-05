*** Settings ***
Documentation     Next-order coupons (BR-V08..V12): issued once at confirm of a qualifying source
...               order with a customer (AC-30, AC-31), redeemable only through the apply phase
...               (AC-32), revoked when the source is cancelled unused (AC-33) and blocking that
...               cancel once used elsewhere (AC-34).
Resource          resources/sales_promotion.resource
Suite Setup       Next Order Promotion
Suite Teardown    Archive Suite Promotions
Test Tags         sales    sales_promotion    next_order_coupon


*** Test Cases ***
A Source Order Without A Customer Confirms But Issues No Coupon
    [Documentation]    AC-30: the quote says why, and the confirm still succeeds.
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}
    ${warning}=    Items Where    ${quote}[warnings]    promotion_id    ${NEXT_ID}
    ${warning}=    Items Where    ${warning}    code    coupon_needs_customer
    Should Not Be Empty    ${warning}
    Confirm Order    ${order}
    ${issued}=    Issuances For    ${order}    ${NEXT_ID}
    Should Be Empty    ${issued}

A Source Order With A Customer Gets One Coupon And No Reward
    [Documentation]    AC-31: the coupon is bound to the customer and the source order; the
    ...    promotion's reward never applies on the source order itself.
    ${order}=    Create Promotion Order
    Assign Customer    ${order}    ${PROMO_CUSTOMER_ID}
    ${quote}=    Quote Of    ${order}
    Promotion Should Not Apply    ${quote}    ${NEXT_ID}
    Confirm Order    ${order}
    ${issued}=    Issuances For    ${order}    ${NEXT_ID}
    Length Should Be    ${issued}    1
    ${voucher}=    Read Voucher    ${issued}[0][voucher_id]
    Should Be Equal    ${voucher}[customer_id]    ${PROMO_CUSTOMER_ID}
    Should Be Equal    ${voucher}[source_order_id]    ${order}
    Should Be Equal    ${voucher}[voucher_type]    single_use
    Confirm Order    ${order}
    ${again}=    Issuances For    ${order}    ${NEXT_ID}
    Length Should Be    ${again}    1
    Set Suite Variable    ${SOURCE_ORDER}    ${order}
    Set Suite Variable    ${COUPON}    ${voucher}

The Coupon Is Not Usable On Its Own Source Order
    [Documentation]    BR-V09 (the source is confirmed, so any action refuses it anyway).
    ${resp}=    Order Action    ${SOURCE_ORDER}    apply_voucher    code=${COUPON}[code]
    Response Status Should Be    ${resp}    400

A Later Order Redeems The Coupon
    ${order}=    Create Promotion Order
    Assign Customer    ${order}    ${PROMO_CUSTOMER_ID}
    ${quote}=    Quote Of    ${order}    apply_voucher    code=${COUPON}[code]
    Program Should Be    ${quote}    ${NEXT_ID}    applied
    Confirm Order    ${order}
    ${voucher}=    Read Voucher    ${COUPON}[id]
    Should Be Equal As Integers    ${voucher}[used_count]    1
    Set Suite Variable    ${USING_ORDER}    ${order}

Cancelling A Source Whose Coupon Was Used Is Refused
    [Documentation]    AC-34: the refusal names the order to handle first.
    ${resp}=    Order Action    ${SOURCE_ORDER}    cancel    reason=robot
    ${item}=    Promotion Error Should Be    ${resp}    promotion.source_order_voucher_used
    Should Be Equal    ${item}[vars][order_id]    ${USING_ORDER}
    ${stored}=    Read Sales Order    ${SOURCE_ORDER}
    Should Be Equal    ${stored}[status]    confirmed

Cancelling A Source Whose Coupon Is Unused Revokes It
    [Documentation]    AC-33: the issuance history stays.
    ${order}=    Create Promotion Order
    Assign Customer    ${order}    ${PROMO_CUSTOMER_ID}
    Confirm Order    ${order}
    ${issued}=    Issuances For    ${order}    ${NEXT_ID}
    Length Should Be    ${issued}    1
    ${resp}=    Order Action    ${order}    cancel    reason=robot
    Response Status Should Be    ${resp}    200
    ${voucher}=    Read Voucher    ${issued}[0][voucher_id]
    Should Be Equal    ${voucher}[status]    revoked
    Should Be Equal    ${voucher}[revocation_reason]    source_order_cancelled
    ${after}=    Issuances For    ${order}    ${NEXT_ID}
    Length Should Be    ${after}    1

A Coupon Whose Apply Phase Is Not Met Gives Nothing And Stays Unused
    [Documentation]    AC-32.
    ${id}=    Create Promotion    next_order_coupon    next_voucher_validity_days=${30}
    Add Rule    ${id}    phase=issue
    Add Rule    ${id}    min_amount=999999999
    Add Product Discount    ${id}    1000
    Link Scope    ${id}
    Activate Promotion    ${id}
    ${source}=    Create Promotion Order
    Assign Customer    ${source}    ${PROMO_CUSTOMER_ID}
    Confirm Order    ${source}
    ${issued}=    Issuances For    ${source}    ${id}
    Length Should Be    ${issued}    1
    ${voucher}=    Read Voucher    ${issued}[0][voucher_id]
    ${order}=    Create Promotion Order
    Assign Customer    ${order}    ${PROMO_CUSTOMER_ID}
    ${quote}=    Quote Of    ${order}    apply_voucher    code=${voucher}[code]
    Program Should Be    ${quote}    ${id}    not_eligible    condition_not_met
    Confirm Order    ${order}
    ${after}=    Read Voucher    ${voucher}[id]
    Should Be Equal As Integers    ${after}[used_count]    0


*** Keywords ***
Next Order Promotion
    Create Authorized API Session
    Ensure Promotion Inputs
    ${id}=    Create Promotion    next_order_coupon    next_voucher_validity_days=${30}
    Add Rule    ${id}    phase=issue
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    Link Scope    ${id}
    Activate Promotion    ${id}
    Set Suite Variable    ${NEXT_ID}    ${id}

Issuances For
    [Arguments]    ${order_id}    ${promotion_id}
    ${issued}=    Issuances Of Order    ${order_id}
    ${mine}=    Items Where    ${issued}    promotion_id    ${promotion_id}
    RETURN    ${mine}
