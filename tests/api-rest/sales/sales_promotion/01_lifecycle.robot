*** Settings ***
Documentation     Promotion lifecycle: Activate refuses a promotion without both scope lists
...               (AC-03, BR-S02), an active promotion is locked for edits (BR-L02), and
...               program_type/currency cannot change once a voucher exists (AC-41).
Resource          resources/sales_promotion.resource
Suite Teardown    Archive Suite Promotions
Test Tags         sales    sales_promotion    lifecycle


*** Test Cases ***
Activate Refuses A Promotion With No Channel And No Point
    [Documentation]    AC-03: an empty channel or point list is never "everywhere".
    ${id}=    Create Promotion    automatic
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    ${resp}=    Promotion Action    ${id}    activate
    Response Status Should Be    ${resp}    400
    ${fields}=    Evaluate    sorted(e['field'] for e in $resp.json() if e['key'] == 'sales:promotion.activation_invalid')
    Should Contain    ${fields}    sales_channel_ids
    Should Contain    ${fields}    sales_point_ids
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal    ${promotion}[status]    draft

Activate Refuses A Promotion With A Channel But No Point
    ${id}=    Create Promotion    automatic
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    ${resp}=    POST On Session    api    ${PROMOTION_API}/${id}/manage_channels
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'add': [$PROMO_CHANNEL_ID]} }}
    Response Status Should Be    ${resp}    200
    ${resp}=    Promotion Action    ${id}    activate
    ${item}=    Promotion Error Should Be    ${resp}    promotion.activation_invalid
    ${fields}=    Evaluate    [e['field'] for e in $resp.json()]
    Should Contain    ${fields}    sales_point_ids
    Should Not Contain    ${fields}    sales_channel_ids

Activate Succeeds Once Both Lists Are Linked
    ${id}=    Create Promotion    automatic
    Add Rule    ${id}
    Add Product Discount    ${id}    1000
    Link Scope    ${id}
    Activate Promotion    ${id}
    ${promotion}=    Read Promotion    ${id}
    Should Be Equal    ${promotion}[status]    active
    Should Match Regexp    ${promotion}[reference]    ^KM\\d+$
    Set Suite Variable    ${ACTIVE_ID}    ${id}

An Active Promotion Refuses Configuration Edits
    [Documentation]    BR-L02: suspend first.
    ${promotion}=    Read Promotion    ${ACTIVE_ID}
    ${resp}=    PATCH On Session    api    ${PROMOTION_API}/${ACTIVE_ID}
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'etag': $promotion['etag'], 'name': 'Renamed while active'} }}
    ...    expected_status=any
    Promotion Error Should Be    ${resp}    promotion.config_locked

Program Type Cannot Change Once A Voucher Exists
    [Documentation]    AC-41.
    ${id}=    Active Discount Promotion    coupon
    Generate Vouchers    ${id}    1
    ${resp}=    Promotion Action    ${id}    suspend
    Response Status Should Be    ${resp}    200
    ${promotion}=    Read Promotion    ${id}
    ${resp}=    PATCH On Session    api    ${PROMOTION_API}/${id}
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'etag': $promotion['etag'], 'program_type': 'discount_code'} }}
    ...    expected_status=any
    Promotion Error Should Be    ${resp}    promotion.immutable_after_use
    ${after}=    Read Promotion    ${id}
    Should Be Equal    ${after}[program_type]    coupon

Automatic And Buy X Get Y Promotions Take No Voucher
    [Documentation]    BR-V01.
    ${resp}=    POST On Session    api    ${VOUCHER_API}/generate_batch
    ...    json=${{ {'org_id': $SALES_ORG_ID, 'promotion_id': $ACTIVE_ID, 'quantity': 1} }}    expected_status=any
    Response Status Should Be    ${resp}    400
