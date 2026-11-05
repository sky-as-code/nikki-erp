*** Settings ***
Documentation     Scope and voucher entry on a draft: channel AND point (AC-02, AC-04, AC-05),
...               code normalising and the shared "invalid" answer (AC-20, AC-21), customer and
...               scope mismatch (AC-22), validity boundaries (AC-23), one voucher per promotion
...               (AC-18), preview consumes nothing (AC-24) and forged amounts are ignored (AC-42).
Resource          resources/sales_promotion.resource
Suite Teardown    Archive Suite Promotions
Test Tags         sales    sales_promotion    voucher


*** Test Cases ***
A Promotion Linked To The Channel But Another Point Does Not Apply
    [Documentation]    AC-02 / AC-04: points are explicit links; the order's point is not one.
    ${id}=    Create Promotion    automatic
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    Link Scope    ${id}    channel_id=${PROMO_CHANNEL_ID}    point_id=${OTHER_POINT_ID}
    Activate Promotion    ${id}
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}
    Promotion Should Not Apply    ${quote}    ${id}
    Archive Promotion    ${id}

A Promotion Linked To The Point But Another Channel Does Not Apply
    [Documentation]    AC-02: a point never implies its channel.
    ${id}=    Create Promotion    automatic
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    Link Scope    ${id}    channel_id=${OTHER_CHANNEL_ID}    point_id=${PROMO_POINT_ID}
    Activate Promotion    ${id}
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}
    Promotion Should Not Apply    ${quote}    ${id}
    Archive Promotion    ${id}

A Promotion Linked To Both Applies
    ${id}=    Active Discount Promotion    automatic    amount=1000
    ${order}=    Create Promotion Order
    ${quote}=    Quote Of    ${order}
    ${program}=    Program Should Be    ${quote}    ${id}    applied
    Should Be Equal    ${program}[discount_untaxed_amount]    1000
    Archive Promotion    ${id}

A Lowercase Code With Outer Spaces Matches The Normalised Code
    [Documentation]    AC-20.
    ${id}=    Active Discount Promotion    discount_code    amount=2000
    Set Suite Variable    ${CODE_PROMOTION_ID}    ${id}
    ${vouchers}=    Generate Vouchers    ${id}    2
    Set Suite Variable    ${CODE_VOUCHERS}    ${vouchers}
    ${order}=    Create Promotion Order
    ${lower}=    Evaluate    $vouchers[0]['code'].lower()
    ${code}=    Set Variable    ${SPACE * 2}${lower}${SPACE}
    ${quote}=    Quote Of    ${order}    apply_voucher    code=${code}
    ${program}=    Program Should Be    ${quote}    ${id}    applied
    Should Be Equal    ${program}[voucher_code]    ${vouchers}[0][code]
    Set Suite Variable    ${CODE_ORDER}    ${order}

A Space Inside The Code Is Not Removed
    ${code}=    Evaluate    $CODE_VOUCHERS[0]['code'][:3] + ' ' + $CODE_VOUCHERS[0]['code'][3:]
    ${resp}=    Order Action    ${CODE_ORDER}    apply_voucher    code=${code}
    Promotion Error Should Be    ${resp}    promotion.invalid_voucher

An Unknown Code Answers Invalid Voucher
    [Documentation]    AC-21: unknown, other-org and archived codes share one answer that names
    ...    neither the promotion nor the customer.
    ${suffix}=    Unique Suffix
    ${resp}=    Order Action    ${CODE_ORDER}    apply_voucher    code=NOPE-${suffix.upper()}
    ${item}=    Promotion Error Should Be    ${resp}    promotion.invalid_voucher
    Should Be Equal    ${item}[field]    code
    Should Not Contain    ${resp.text}    ${CODE_PROMOTION_ID}

A Code Of Another Organization Answers The Same Invalid Voucher
    [Documentation]    AC-21: a code that exists in another org is never found from this one, and
    ...    the answer is the one an unknown code gets.
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 10, 'fields': 'id'} }}
    Response Status Should Be    ${resp}    200
    ${others}=    Evaluate    [o['id'] for o in $resp.json()['items']]
    Remove Values From List    ${others}    ${SALES_ORG_ID}
    IF    not $others    Skip    A second organization is needed
    ${other_org}=    Set Variable    ${others}[0]
    ${name}=    Unique Display Name    Robot other-org promotion
    ${resp}=    POST On Session    api    ${PROMOTION_API}    expected_status=any
    ...    json=${{ {'org_id': $other_org, 'name': $name, 'program_type': 'coupon', 'currency_id': $PROMO_CURRENCY_ID, 'starts_at': $PROMOTION_STARTS_AT} }}
    IF    $resp.status_code != 201    Skip    Cannot create a promotion in ${other_org}: ${resp.text}
    ${promotion_id}=    Set Variable    ${resp.json()}[id]
    ${resp}=    POST On Session    api    ${VOUCHER_API}/generate_batch    expected_status=any
    ...    json=${{ {'org_id': $other_org, 'promotion_id': $promotion_id, 'quantity': 1, 'prefix': 'RB'} }}
    IF    $resp.status_code != 200    Skip    Cannot generate a voucher in ${other_org}: ${resp.text}
    ${code}=    Set Variable    ${resp.json()}[vouchers][0][code]
    ${resp}=    Order Action    ${CODE_ORDER}    apply_voucher    code=${code}
    ${item}=    Promotion Error Should Be    ${resp}    promotion.invalid_voucher
    Should Not Contain    ${resp.text}    ${promotion_id}

A Second Code Of The Same Promotion Must Replace The First
    [Documentation]    AC-18: never two benefits from one promotion.
    ${resp}=    Order Action    ${CODE_ORDER}    apply_voucher    code=${CODE_VOUCHERS}[1][code]
    Promotion Error Should Be    ${resp}    promotion.voucher_replace_required
    ${quote}=    Quote Of    ${CODE_ORDER}    apply_voucher    code=${CODE_VOUCHERS}[1][code]    replace_existing=${True}
    ${program}=    Program Should Be    ${quote}    ${CODE_PROMOTION_ID}    applied
    Should Be Equal    ${program}[voucher_code]    ${CODE_VOUCHERS}[1][code]
    ${mine}=    Items Where    ${quote}[programs]    promotion_id    ${CODE_PROMOTION_ID}
    Length Should Be    ${mine}    1

Previewing Many Times Consumes Nothing
    [Documentation]    AC-24 / BR-V07: no use is taken or held before confirm.
    FOR    ${index}    IN RANGE    3
        Quote Of    ${CODE_ORDER}
    END
    ${voucher}=    Read Voucher    ${CODE_VOUCHERS}[1][id]
    Should Be Equal As Integers    ${voucher}[used_count]    0
    ${promotion}=    Read Promotion    ${CODE_PROMOTION_ID}
    Should Be Equal As Integers    ${promotion}[used_count]    0

Forged Amounts In The Request Are Ignored
    [Documentation]    AC-42 / BR-T01: the server computes the benefit itself.
    ${quote}=    Quote Of    ${CODE_ORDER}    preview_promotions    discount_untaxed_amount=999999
    ...    discount_total=999999    gift_quantity=50
    ${program}=    Program Should Be    ${quote}    ${CODE_PROMOTION_ID}    applied
    Should Be Equal    ${program}[discount_untaxed_amount]    2000

Removing The Voucher Removes Its Benefit
    ${quote}=    Quote Of    ${CODE_ORDER}    remove_voucher    promotion_id=${CODE_PROMOTION_ID}
    Promotion Should Not Apply    ${quote}    ${CODE_PROMOTION_ID}

A Voucher Of Another Customer Is Refused
    [Documentation]    AC-22 / BR-V04.
    ${vouchers}=    Generate Vouchers    ${CODE_PROMOTION_ID}    1    customer_id=${OTHER_CUSTOMER_ID}
    ${order}=    Create Promotion Order
    Assign Customer    ${order}    ${PROMO_CUSTOMER_ID}
    ${resp}=    Order Action    ${order}    apply_voucher    code=${vouchers}[0][code]
    Promotion Error Should Be    ${resp}    promotion.voucher_customer_mismatch

A Voucher Out Of The Order's Scope Is Refused
    [Documentation]    AC-22: a valid code never bypasses the channel and point scope.
    ${id}=    Create Promotion    coupon
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    Link Scope    ${id}    channel_id=${OTHER_CHANNEL_ID}    point_id=${OTHER_POINT_ID}
    Activate Promotion    ${id}
    ${vouchers}=    Generate Vouchers    ${id}    1
    ${order}=    Create Promotion Order
    ${resp}=    Order Action    ${order}    apply_voucher    code=${vouchers}[0][code]
    Promotion Error Should Be    ${resp}    promotion.scope_mismatch

A Voucher Before Its Start Or After Its End Is Not Available
    [Documentation]    AC-23: the window is [valid_from, expires_at).
    ${future}=    Evaluate    (__import__('datetime').datetime.utcnow() + __import__('datetime').timedelta(days=2)).strftime('%Y-%m-%dT%H:%M:%SZ')
    ${past}=    Evaluate    (__import__('datetime').datetime.utcnow() - __import__('datetime').timedelta(days=1)).strftime('%Y-%m-%dT%H:%M:%SZ')
    ${early}=    Generate Vouchers    ${CODE_PROMOTION_ID}    1    valid_from=${future}
    ${late}=    Generate Vouchers    ${CODE_PROMOTION_ID}    1    valid_from=2026-01-01T00:00:00Z    expires_at=${past}
    ${order}=    Create Promotion Order
    ${resp}=    Order Action    ${order}    apply_voucher    code=${early}[0][code]
    Promotion Error Should Be    ${resp}    promotion.not_available
    ${resp}=    Order Action    ${order}    apply_voucher    code=${late}[0][code]
    Promotion Error Should Be    ${resp}    promotion.not_available

An Order Without A Sales Point Gets No Promotion
    [Documentation]    AC-05: no context, no benefit; the quote says why.
    ${line}=    Create Dictionary    product_variant_id=${SALES_VARIANT_ID}    uom_id=${SALES_UOM_ID}
    ...    quantity=1    unit_price=50000
    ${resp}=    POST On Session    api    ${SALES_CREATE_ORDER_API}    expected_status=any
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'sales_channel_code': $PROMO_CHANNEL_CODE, 'currency_code': 'VND', 'lines': [$line]} }}
    IF    $resp.status_code != 200    Skip    create_order requires a sales point, so AC-05 cannot arise here
    ${quote}=    Quote Of    ${resp.json()}[sales_order_id]
    ${codes}=    Evaluate    [w['code'] for w in $quote['warnings']]
    Should Contain    ${codes}    context_required
    ${applied}=    Evaluate    [p for p in $quote['programs'] if p['outcome'] == 'applied']
    Should Be Empty    ${applied}
