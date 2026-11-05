*** Settings ***
Documentation     Helpers for the Sales promotion and voucher suites (BR_Sales_Promotion_Voucher).
...               Every suite builds its own promotions through the API, so runs are repeatable and
...               independent of seed data, and archives them again in its teardown: an active
...               automatic promotion left behind would apply to every later order at that point.
...               Discount rewards use the specific_products scope where possible, so they never
...               compete with another order-scope promotion of the environment (BR-R14).
Library           Collections
Library           RequestsLibrary
Resource          ${CURDIR}/sales.resource


*** Variables ***
${PROMOTION_API}            /v1/sales/sales_promotion
${PROMOTION_RULE_API}       /v1/sales/sales_promotion_rule
${PROMOTION_REWARD_API}     /v1/sales/sales_promotion_reward
${VOUCHER_API}              /v1/sales/sales_voucher
${APPLICATION_API}          /v1/sales/sales_promotion_application
${ISSUANCE_API}             /v1/sales/sales_voucher_issuance
${PROMOTION_STARTS_AT}      2026-01-01T00:00:00Z


*** Keywords ***
Items Where
    [Documentation]    The items whose field equals value. A loop rather than a comprehension: an
    ...    Evaluate comprehension cannot see Robot variables in its condition.
    [Arguments]    ${items}    ${field}    ${value}
    ${matching}=    Create List
    FOR    ${item}    IN    @{items}
        IF    $item.get($field) == $value    Append To List    ${matching}    ${item}
    END
    RETURN    ${matching}

Item Where
    [Documentation]    The first item whose field equals value, or None.
    [Arguments]    ${items}    ${field}    ${value}
    ${matching}=    Items Where    ${items}    ${field}    ${value}
    IF    not $matching    RETURN    ${None}
    RETURN    ${matching}[0]

Ensure Promotion Inputs
    [Documentation]    Resolves once: two active sales points of DIFFERENT channels (the order is
    ...    raised at the first), a sellable variant, a unit, the VND currency and a party.
    ${done}=    Get Variable Value    ${PROMO_POINT_ID}    ${EMPTY}
    IF    $done    RETURN
    Ensure Sales Order Inputs
    ${json}=    Evaluate    __import__('json')
    ${resp}=    GET On Session    api    ${SALES_POINT_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'size': 50, 'fields': 'id,sales_channel_id', 'graph': $json.dumps({'and': [{'if': ['status', '=', 'active']}, {'if': ['is_archived', '=', False]}]})} }}
    Response Status Should Be    ${resp}    200
    # The last points rather than the first: environments seed their kiosk points first, and a
    # kiosk confirm holds stock the suites have no reason to touch.
    ${points}=    Evaluate    list(reversed($resp.json()['items']))
    IF    len($points) < 2    Skip    Two active sales points are needed in org ${SALES_ORG_ID}
    ${first}=    Set Variable    ${points}[0]
    ${other}=    Set Variable    ${None}
    FOR    ${point}    IN    @{points}
        IF    $point['sales_channel_id'] != $first['sales_channel_id']
            ${other}=    Set Variable    ${point}
            BREAK
        END
    END
    IF    $other is None    Skip    Two sales points of different channels are needed
    ${resp}=    GET On Session    api    ${SALES_CHANNEL_API}/${first}[sales_channel_id]    params=${{ {'org_id': $SALES_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Set Global Variable    ${PROMO_POINT_ID}    ${first}[id]
    Set Global Variable    ${PROMO_CHANNEL_ID}    ${first}[sales_channel_id]
    Set Global Variable    ${PROMO_CHANNEL_CODE}    ${resp.json()}[item][code]
    Set Global Variable    ${OTHER_POINT_ID}    ${other}[id]
    Set Global Variable    ${OTHER_CHANNEL_ID}    ${other}[sales_channel_id]
    ${resp}=    GET On Session    api    /v1/essential/essential_currency
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'size': 1, 'fields': 'id', 'graph': $json.dumps({'if': ['code', '=', 'VND']})} }}
    ${currency}=    First Item Or Skip    ${resp}    VND currency
    Set Global Variable    ${PROMO_CURRENCY_ID}    ${currency}[id]
    ${resp}=    GET On Session    api    /v1/contacts/contacts_party    params=${{ {'org_id': $SALES_ORG_ID, 'size': 2, 'fields': 'id'} }}
    Response Status Should Be    ${resp}    200
    IF    len($resp.json()['items']) < 2    Skip    Two contacts parties are needed
    Set Global Variable    ${PROMO_CUSTOMER_ID}    ${resp.json()}[items][0][id]
    Set Global Variable    ${OTHER_CUSTOMER_ID}    ${resp.json()}[items][1][id]

Remember Promotion
    [Arguments]    ${promotion_id}
    ${ids}=    Get Variable Value    ${SUITE_PROMOTION_IDS}    ${None}
    IF    $ids is None
        ${ids}=    Create List
        Set Suite Variable    ${SUITE_PROMOTION_IDS}    ${ids}
    END
    Append To List    ${SUITE_PROMOTION_IDS}    ${promotion_id}

Archive Suite Promotions
    [Documentation]    Teardown: archives every promotion the suite created, so none stays live.
    ${ids}=    Get Variable Value    ${SUITE_PROMOTION_IDS}    ${None}
    IF    $ids is None    RETURN
    FOR    ${id}    IN    @{ids}
        Archive Promotion    ${id}
    END

Archive Promotion
    [Arguments]    ${promotion_id}
    ${promotion}=    Read Promotion    ${promotion_id}
    IF    $promotion['is_archived']    RETURN
    ${resp}=    POST On Session    api    ${PROMOTION_API}/${promotion_id}/archived
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'etag': $promotion['etag'], 'is_archived': True} }}    expected_status=any
    Response Status Should Be    ${resp}    200

Read Promotion
    [Arguments]    ${promotion_id}
    ${resp}=    GET On Session    api    ${PROMOTION_API}/${promotion_id}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'include_archived': True} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Create Promotion
    [Documentation]    A draft promotion in VND, running since 2026-01-01; returns its id.
    [Arguments]    ${program_type}    &{extra}
    Ensure Promotion Inputs
    ${name}=    Unique Display Name    Robot promotion
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    name=${name}    program_type=${program_type}
    ...    currency_id=${PROMO_CURRENCY_ID}    starts_at=${PROMOTION_STARTS_AT}    &{extra}
    ${resp}=    POST On Session    api    ${PROMOTION_API}    json=${body}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Remember Promotion    ${id}
    RETURN    ${id}

Add Rule
    [Arguments]    ${promotion_id}    ${phase}=apply    ${min_quantity}=0    ${min_amount}=0    &{extra}
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    promotion_id=${promotion_id}    phase=${phase}
    ...    name=Robot rule    min_quantity=${min_quantity}    min_amount=${min_amount}
    ...    amount_tax_basis=tax_excluded    &{extra}
    ${resp}=    POST On Session    api    ${PROMOTION_RULE_API}    json=${body}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Add Product Discount
    [Documentation]    A fixed-amount discount on the order's variant (specific_products scope).
    [Arguments]    ${promotion_id}    ${amount}
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    promotion_id=${promotion_id}    reward_type=discount
    ...    order_description=Robot discount    discount_method=fixed_amount    discount_value=${amount}
    ...    discount_scope=specific_products    max_discount_amount=0
    ${resp}=    POST On Session    api    ${PROMOTION_REWARD_API}    json=${body}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${product}=    Create Dictionary    product_variant_id=${SALES_VARIANT_ID}
    ${resp}=    POST On Session    api    ${PROMOTION_REWARD_API}/${id}/manage_products
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'add': [$product]} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${id}

Add Gift Reward
    [Arguments]    ${promotion_id}    ${quantity}=1
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    promotion_id=${promotion_id}    reward_type=free_product
    ...    order_description=Robot gift    gift_product_variant_id=${SALES_VARIANT_ID}    gift_quantity=${quantity}
    ...    gift_uom_id=${SALES_UOM_ID}
    ${resp}=    POST On Session    api    ${PROMOTION_REWARD_API}    json=${body}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Link Scope
    [Arguments]    ${promotion_id}    ${channel_id}=${PROMO_CHANNEL_ID}    ${point_id}=${PROMO_POINT_ID}
    ${resp}=    POST On Session    api    ${PROMOTION_API}/${promotion_id}/manage_channels
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'add': [$channel_id]} }}
    Response Status Should Be    ${resp}    200
    ${resp}=    POST On Session    api    ${PROMOTION_API}/${promotion_id}/manage_points
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'add': [$point_id]} }}
    Response Status Should Be    ${resp}    200

Promotion Action
    [Arguments]    ${promotion_id}    ${action}
    ${resp}=    POST On Session    api    ${PROMOTION_API}/${promotion_id}/${action}
    ...    json=${{ {'org_id': $SALES_ORG_ID} }}    expected_status=any
    RETURN    ${resp}

Activate Promotion
    [Arguments]    ${promotion_id}
    ${resp}=    Promotion Action    ${promotion_id}    activate
    Response Status Should Be    ${resp}    200

Active Discount Promotion
    [Documentation]    A ready-to-use promotion: one apply rule, one product discount, scoped to the
    ...    order point, active. Returns its id.
    [Arguments]    ${program_type}    ${amount}=1000    ${min_amount}=0    &{extra}
    ${id}=    Create Promotion    ${program_type}    &{extra}
    Add Rule    ${id}    min_amount=${min_amount}
    Add Product Discount    ${id}    ${amount}
    Link Scope    ${id}
    Activate Promotion    ${id}
    RETURN    ${id}

Generate Vouchers
    [Arguments]    ${promotion_id}    ${quantity}=1    &{extra}
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    promotion_id=${promotion_id}    quantity=${{ int($quantity) }}
    ...    prefix=RB    &{extra}
    ${resp}=    POST On Session    api    ${VOUCHER_API}/generate_batch    json=${body}    expected_status=any
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[vouchers]

Read Voucher
    [Arguments]    ${voucher_id}
    ${resp}=    GET On Session    api    ${VOUCHER_API}/${voucher_id}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'include_archived': True} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Create Promotion Order
    [Documentation]    A draft at the promotion point with one line of the variant.
    [Arguments]    ${quantity}=2    ${point_id}=${PROMO_POINT_ID}    ${channel_code}=${PROMO_CHANNEL_CODE}
    Ensure Promotion Inputs
    ${line}=    Create Dictionary    product_variant_id=${SALES_VARIANT_ID}    uom_id=${SALES_UOM_ID}
    ...    quantity=${quantity}    unit_price=50000
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    sales_channel_code=${channel_code}
    ...    sales_point_id=${point_id}    currency_code=VND    lines=${{ [$line] }}
    ${resp}=    POST On Session    api    ${SALES_CREATE_ORDER_API}    json=${body}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[sales_order_id]

Order Action
    [Documentation]    POSTs an order action; returns the raw response (callers assert).
    [Arguments]    ${order_id}    ${action}    &{body}
    Set To Dictionary    ${body}    org_id=${SALES_ORG_ID}
    ${resp}=    POST On Session    api    ${SALES_ORDER_API}/${order_id}/${action}    json=${body}    expected_status=any
    RETURN    ${resp}

Quote Of
    [Documentation]    Runs a draft promotion action that must succeed and returns its PromotionQuote.
    [Arguments]    ${order_id}    ${action}=preview_promotions    &{body}
    ${resp}=    Order Action    ${order_id}    ${action}    &{body}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[calculation_reference]
    RETURN    ${resp.json()}

Program Of
    [Documentation]    The quote's entry for a promotion, or None.
    [Arguments]    ${quote}    ${promotion_id}
    ${program}=    Item Where    ${quote}[programs]    promotion_id    ${promotion_id}
    RETURN    ${program}

Program Should Be
    [Arguments]    ${quote}    ${promotion_id}    ${outcome}    ${reason}=${None}
    ${program}=    Program Of    ${quote}    ${promotion_id}
    Should Not Be Equal    ${program}    ${None}    msg=Promotion ${promotion_id} is not in the quote
    Should Be Equal    ${program}[outcome]    ${outcome}
    IF    $reason is not None    Should Be Equal    ${program.get('reason_code')}    ${reason}
    RETURN    ${program}

Promotion Should Not Apply
    [Arguments]    ${quote}    ${promotion_id}
    ${program}=    Program Of    ${quote}    ${promotion_id}
    IF    $program is not None    Should Not Be Equal    ${program}[outcome]    applied

Promotion Error Should Be
    [Documentation]    A business refusal keyed sales:<key>; returns the error item.
    [Arguments]    ${resp}    ${key}
    Response Status Should Be    ${resp}    400
    ${item}=    Item Where    ${resp.json()}    key    sales:${key}
    Should Not Be Equal    ${item}    ${None}    msg=No ${key} in ${resp.text}
    RETURN    ${item}

Confirm Order
    [Arguments]    ${order_id}    &{body}
    ${resp}=    Order Action    ${order_id}    confirm    &{body}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}

Assign Customer
    [Arguments]    ${order_id}    ${party_id}
    ${resp}=    Order Action    ${order_id}    assign_sold_to_party    sold_to_party_id=${party_id}
    Response Status Should Be    ${resp}    200

Applications Of Order
    [Arguments]    ${order_id}
    ${json}=    Evaluate    __import__('json')
    ${resp}=    GET On Session    api    ${APPLICATION_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'fields': 'id,status,promotion_id,voucher_id', 'graph': $json.dumps({'if': ['order_id', '=', $order_id]})} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items]

Issuances Of Order
    [Arguments]    ${order_id}
    ${json}=    Evaluate    __import__('json')
    ${resp}=    GET On Session    api    ${ISSUANCE_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'fields': 'id,voucher_id,promotion_id', 'graph': $json.dumps({'if': ['source_order_id', '=', $order_id]})} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items]
