*** Settings ***
Documentation     Reading purchase orders: by id, by existence, and by search.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    read


*** Variables ***
${ORDER_SCHEMA}    ${PURCHASE_SCHEMA_DIR}/purchase_order.json


*** Test Cases ***
Get Purchase Order By Id
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    ...    params=${{ {'org_id': $PURCHASE_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[item][id]    ${PURCHASE_ORDER_ID}

Get Unknown Purchase Order Is Not Found
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/01JZZZZZZZZZZZZZZZZZZZZZZZ
    ...    params=${{ {'org_id': $PURCHASE_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Purchase Order Exists
    Ensure Purchase Order Under Test
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/exists
    ...    json=${{ {'org_id': $PURCHASE_ORG_ID, 'ids': [$PURCHASE_ORDER_ID, $NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=1

Search Purchase Orders
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'org_id': $PURCHASE_ORG_ID, 'size': 20} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=20    page=0

Search By Status Returns Only That Status
    [Documentation]    The status filter is what the frontend's RFQ and PO tabs are built on —
    ...    one resource, two views (PUR-R1) — so a filter that leaked another status would put
    ...    confirmed orders on the quotations screen.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'org_id': $PURCHASE_ORG_ID, 'size': 50, 'graph': '{"if":["status","=","rfq"]}'} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=50    page=0
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[status]    rfq
    END

Search By Vendor Returns Only That Vendor
    Ensure Purchase Order Under Test
    ${graph}=    Set Variable    {"if":["vendor_id","=","${PURCHASE_VENDOR_ID}"]}
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'org_id': $PURCHASE_ORG_ID, 'size': 50, 'graph': $graph, 'fields': 'id,vendor_id'} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=50    page=0
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[vendor_id]    ${PURCHASE_VENDOR_ID}
    END
