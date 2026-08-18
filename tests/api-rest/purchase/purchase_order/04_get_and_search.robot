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
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[id]    ${PURCHASE_ORDER_ID}

Get Unknown Purchase Order Is Not Found
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/01JZZZZZZZZZZZZZZZZZZZZZZZ    expected_status=any
    Response Should Be Not Found Error    ${resp}

Purchase Order Exists
    Ensure Purchase Order Under Test
    ${existing}=    Create List    ${PURCHASE_ORDER_ID}
    ${missing}=    Create List    01JZZZZZZZZZZZZZZZZZZZZZZZ
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/exists    params=${{ {'id': $PURCHASE_ORDER_ID} }}
    Response Should Be Exists Success    ${resp}    ${existing}    ${missing}

Search Purchase Orders
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'size': 20} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=20    page=0

Search By Status Returns Only That Status
    [Documentation]    The status filter is what the frontend's RFQ and PO tabs are built on —
    ...    one resource, two views (PUR-R1) — so a filter that leaked another status would put
    ...    confirmed orders on the quotations screen.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'size': 50, 'graph': '{"if":["status","eq","rfq"]}'} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=50    page=0
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[status]    rfq
    END

Search By Vendor Returns Only That Vendor
    Ensure Purchase Order Under Test
    ${graph}=    Set Variable    {"if":["vendor_id","eq","${PURCHASE_VENDOR_ID}"]}
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}    params=${{ {'size': 50, 'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${ORDER_SCHEMA}    size=50    page=0
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[vendor_id]    ${PURCHASE_VENDOR_ID}
    END
