*** Settings ***
Documentation     Sales helpers for the order numbering suite. Orders are raised through the
...               create_order action, the path every channel (back office, kiosk, online) takes.
...               The inputs are resolved from the API rather than hardcoded: an active sales point
...               with its channel, a sellable variant and a unit. A missing one skips the test,
...               since nothing here can assume an environment has them.
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${SALES_ORDER_API}          /v1/sales/sales_order
${SALES_CREATE_ORDER_API}   /v1/sales/sales_order/create_order
${SALES_POINT_API}          /v1/sales/sales_point
${SALES_CHANNEL_API}        /v1/sales/sales_channel
# Crockford Base32: digits and uppercase letters without I, L, O and U.
${SALES_CROCKFORD_CHAR}     [0-9A-HJKMNP-TV-Z]


*** Keywords ***
Ensure Sales Org
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${SALES_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${SALES_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[items]
    Set Global Variable    ${SALES_ORG_ID}    ${resp.json()}[items][0][id]

First Item Or Skip
    [Arguments]    ${resp}    ${what}
    Response Status Should Be    ${resp}    200
    IF    not $resp.json()['items']    Skip    No ${what} in org ${SALES_ORG_ID}
    RETURN    ${resp.json()}[items][0]

Ensure Sales Order Inputs
    [Documentation]    Resolves once: SALES_POINT_ID, SALES_CHANNEL_CODE, SALES_VARIANT_ID, SALES_UOM_ID.
    ${done}=    Get Variable Value    ${SALES_VARIANT_ID}    ${EMPTY}
    IF    $done    RETURN
    Ensure Sales Org
    ${json}=    Evaluate    __import__('json')
    ${resp}=    GET On Session    api    ${SALES_POINT_API}
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'size': 1, 'fields': 'id,sales_channel_id', 'graph': $json.dumps({'and': [{'if': ['status', '=', 'active']}, {'if': ['is_archived', '=', False]}]})} }}
    ${point}=    First Item Or Skip    ${resp}    active sales point
    ${resp}=    GET On Session    api    ${SALES_CHANNEL_API}/${point}[sales_channel_id]
    ...    params=${{ {'org_id': $SALES_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${resp_variant}=    GET On Session    api    /v1/inventory/inventory_product_variant
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'size': 1, 'fields': 'id', 'graph': $json.dumps({'and': [{'if': ['status', '=', 'active']}, {'if': ['variant_kind', '!=', 'base']}, {'if': ['is_archived', '=', False]}]})} }}
    ${variant}=    First Item Or Skip    ${resp_variant}    sellable product variant
    ${resp_uom}=    GET On Session    api    /v1/essential/essential_uom
    ...    params=${{ {'org_id': $SALES_ORG_ID, 'size': 1, 'fields': 'id'} }}
    ${uom}=    First Item Or Skip    ${resp_uom}    unit of measure
    Set Global Variable    ${SALES_POINT_ID}    ${point}[id]
    Set Global Variable    ${SALES_CHANNEL_CODE}    ${resp.json()}[item][code]
    Set Global Variable    ${SALES_UOM_ID}    ${uom}[id]
    Set Global Variable    ${SALES_VARIANT_ID}    ${variant}[id]

Create Sales Order
    [Documentation]    Raises a one-line draft through create_order; returns the response body
    ...    (sales_order_id, order_number, public_reference, already_existed, totals).
    [Arguments]    ${idempotency_key}=${None}
    Ensure Sales Order Inputs
    ${line}=    Create Dictionary    product_variant_id=${SALES_VARIANT_ID}    uom_id=${SALES_UOM_ID}
    ...    quantity=1    unit_price=1000
    ${body}=    Create Dictionary    org_id=${SALES_ORG_ID}    sales_channel_code=${SALES_CHANNEL_CODE}
    ...    sales_point_id=${SALES_POINT_ID}    currency_code=VND    lines=${{ [$line] }}
    IF    $idempotency_key is not None    Set To Dictionary    ${body}    idempotency_key=${idempotency_key}
    ${resp}=    POST On Session    api    ${SALES_CREATE_ORDER_API}    json=${body}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}

Read Sales Order
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${SALES_ORDER_API}/${id}    params=${{ {'org_id': $SALES_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Delete Sales Order Fixture
    [Arguments]    ${id}
    DELETE On Session    api    ${SALES_ORDER_API}/${id}    params=${{ {'org_id': $SALES_ORG_ID} }}
    ...    expected_status=any
