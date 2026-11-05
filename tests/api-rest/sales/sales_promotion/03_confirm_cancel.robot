*** Settings ***
Documentation     Confirm and cancel: usage is taken atomically at confirm (AC-25, AC-26), a lost
...               race leaves the order a draft with nothing consumed and a fresh quote (AC-28,
...               BR-T04), a retried confirm duplicates nothing (AC-27), a changed calculation
...               asks for a requote (AC-29), cancel gives the uses back exactly once (AC-35) and an
...               archived promotion keeps confirmed orders intact (AC-37).
Library           Collections
Resource          resources/sales_promotion.resource
Suite Teardown    Archive Suite Promotions
Test Tags         sales    sales_promotion    confirm


*** Test Cases ***
Confirm Records One Application And Takes One Use
    ${id}=    Active Discount Promotion    coupon    amount=1500
    ${vouchers}=    Generate Vouchers    ${id}    1
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}    apply_voucher    code=${vouchers}[0][code]
    Program Should Be    ${quote}    ${id}    applied
    ${confirmed}=    Confirm Order    ${order}    accepted_calculation_reference=${quote}[calculation_reference]
    Length Should Be    ${confirmed}[promotion_applications]    1
    ${voucher}=    Read Voucher    ${vouchers}[0][id]
    Should Be Equal As Integers    ${voucher}[used_count]    1
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal As Integers    ${promotion}[used_count]    1
    Set Suite Variable    ${CONFIRMED_ORDER}    ${order}
    Set Suite Variable    ${CONFIRMED_PROMOTION}    ${id}
    Set Suite Variable    ${CONFIRMED_VOUCHER}    ${vouchers}[0][id]
    Set Suite Variable    ${CONFIRMED_APPLICATIONS}    ${confirmed}[promotion_applications]

A Retried Confirm Duplicates Nothing
    [Documentation]    AC-27 / BR-T05: the same applications, no second use.
    ${replay}=    Confirm Order    ${CONFIRMED_ORDER}
    Should Be True    ${replay}[already_confirmed]
    Lists Should Be Equal    ${replay}[promotion_applications]    ${CONFIRMED_APPLICATIONS}
    ${applications}=    Applications Of Order    ${CONFIRMED_ORDER}
    Length Should Be    ${applications}    1
    ${promotion}=    Read Promotion    ${CONFIRMED_PROMOTION}
    Should Be Equal As Integers    ${promotion}[used_count]    1

Cancel Gives The Uses Back Exactly Once
    [Documentation]    AC-35: the application is reversed and both counters drop by one; a
    ...    retried cancel is refused and changes nothing.
    ${resp}=    Order Action    ${CONFIRMED_ORDER}    cancel    reason=robot
    Response Status Should Be    ${resp}    200
    ${resp}=    Order Action    ${CONFIRMED_ORDER}    cancel    reason=robot
    Response Status Should Be    ${resp}    400
    ${applications}=    Applications Of Order    ${CONFIRMED_ORDER}
    Should Be Equal    ${applications}[0][status]    reversed
    ${voucher}=    Read Voucher    ${CONFIRMED_VOUCHER}
    Should Be Equal As Integers    ${voucher}[used_count]    0
    ${promotion}=    Read Promotion    ${CONFIRMED_PROMOTION}
    Should Be Equal As Integers    ${promotion}[used_count]    0

A Changed Calculation Asks For A Requote
    [Documentation]    AC-29 / BR-T02: the fresh quote travels in vars.quote; nothing is consumed
    ...    and the order stays a draft until the caller accepts the new calculation.
    ${id}=    Active Discount Promotion    discount_code    amount=1000
    ${vouchers}=    Generate Vouchers    ${id}    1
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}    apply_voucher    code=${vouchers}[0][code]
    Change Discount    ${id}    3000
    ${resp}=    Order Action    ${order}    confirm    accepted_calculation_reference=${quote}[calculation_reference]
    ${item}=    Promotion Error Should Be    ${resp}    promotion.requote_required
    ${fresh}=    Set Variable    ${item}[vars][quote]
    Should Not Be Equal    ${fresh}[calculation_reference]    ${quote}[calculation_reference]
    ${program}=    Program Should Be    ${fresh}    ${id}    applied
    Should Be Equal    ${program}[discount_untaxed_amount]    3000
    ${stored}=    Read Sales Order    ${order}
    Should Be Equal    ${stored}[status]    draft
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal As Integers    ${promotion}[used_count]    0
    Confirm Order    ${order}    accepted_calculation_reference=${fresh}[calculation_reference]

Two Confirms Racing For One Single-Use Voucher Let Only One Win
    [Documentation]    AC-25 / AC-28: the loser stays a draft, consumes nothing and is answered
    ...    with the recomputed quote.
    ${id}=    Active Discount Promotion    coupon    amount=1000
    ${vouchers}=    Generate Vouchers    ${id}    1
    ${first}=    Create Promotion Order
    ${second}=    Create Promotion Order
    Quote Of    ${first}    apply_voucher    code=${vouchers}[0][code]
    Quote Of    ${second}    apply_voucher    code=${vouchers}[0][code]
    ${responses}=    Confirm In Parallel    ${first}    ${second}
    Exactly One Confirm Won    ${responses}    promotion.voucher_usage_exhausted
    ${voucher}=    Read Voucher    ${vouchers}[0][id]
    Should Be Equal As Integers    ${voucher}[used_count]    1
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal As Integers    ${promotion}[used_count]    1

Two Vouchers Racing For The Promotion's Last Use Let Only One Win
    [Documentation]    AC-26: the promotion's own limit is taken atomically too.
    ${id}=    Active Discount Promotion    coupon    amount=1000    usage_limit=${1}
    ${vouchers}=    Generate Vouchers    ${id}    2
    ${first}=    Create Promotion Order
    ${second}=    Create Promotion Order
    Quote Of    ${first}    apply_voucher    code=${vouchers}[0][code]
    Quote Of    ${second}    apply_voucher    code=${vouchers}[1][code]
    ${responses}=    Confirm In Parallel    ${first}    ${second}
    Exactly One Confirm Won    ${responses}    promotion.usage_exhausted
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal As Integers    ${promotion}[used_count]    1

Archiving A Promotion Keeps Confirmed Orders And Stops New Ones
    [Documentation]    AC-37 / BR-L03.
    ${id}=    Active Discount Promotion    automatic    amount=1000
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}
    Program Should Be    ${quote}    ${id}    applied
    Confirm Order    ${order}
    Archive Promotion    ${id}
    ${applications}=    Applications Of Order    ${order}
    ${mine}=    Items Where    ${applications}    promotion_id    ${id}
    Should Be Equal    ${mine}[0][status]    confirmed
    ${draft}=    Create Promotion Order
    ${quote}=    Quote Of    ${draft}
    Promotion Should Not Apply    ${quote}    ${id}


*** Keywords ***
Change Discount
    [Documentation]    Suspend, change the reward amount, activate again: a configuration change.
    [Arguments]    ${promotion_id}    ${amount}
    ${resp}=    Promotion Action    ${promotion_id}    suspend
    Response Status Should Be    ${resp}    200
    ${json}=    Evaluate    __import__('json')
    ${resp}=    GET On Session    api    ${PROMOTION_REWARD_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'fields': 'id,etag', 'graph': $json.dumps({'if': ['promotion_id', '=', $promotion_id]})} }}
    Response Status Should Be    ${resp}    200
    ${reward}=    Set Variable    ${resp.json()}[items][0]
    ${resp}=    PATCH On Session    api    ${PROMOTION_REWARD_API}/${reward}[id]
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'etag': $reward['etag'], 'discount_value': $amount} }}
    Response Status Should Be    ${resp}    200
    Activate Promotion    ${promotion_id}

Confirm In Parallel
    [Documentation]    Two confirms sent at the same moment on the authorised session.
    [Arguments]    ${first}    ${second}
    ${library}=    Get Library Instance    RequestsLibrary
    ${session}=    Call Method    ${library._cache}    switch    api
    ${responses}=    Evaluate
    ...    list(__import__('concurrent.futures', fromlist=['ThreadPoolExecutor']).ThreadPoolExecutor(2).map(lambda order, session=$session, base=$API_HOST + $SALES_ORDER_API, org=$SALES_ORG_ID: session.post(base + '/' + order + '/confirm', json={'org_id': org}, timeout=120), [$first, $second]))
    RETURN    ${responses}

Exactly One Confirm Won
    [Arguments]    ${responses}    ${loser_key}
    ${codes}=    Evaluate    sorted(r.status_code for r in $responses)
    Should Be Equal    ${codes}    ${{ [200, 400] }}    msg=${responses[0].text} | ${responses[1].text}
    ${loser}=    Evaluate    next(r for r in $responses if r.status_code == 400)
    ${item}=    Promotion Error Should Be    ${loser}    ${loser_key}
    Dictionary Should Contain Key    ${item}[vars]    quote
