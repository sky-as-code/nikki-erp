*** Settings ***
Documentation     Order lines, the stored totals, and the unit-of-measure rules.
...               AC-05 and AC-UOM-PUR-01 through AC-UOM-PUR-11.
...
...               PUR-R4 and D8: the three money fields are stored per line and per header, and
...               recomputed on every line write in the SAME transaction. What that buys is an
...               order that still reads correctly years later, when prices and rounding rules
...               have both moved on. What it costs is that a stored value can disagree with the
...               lines — and where it does, the lines win, because they are what a reader can
...               verify by adding up.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order_line    totals


*** Test Cases ***
Add A Line To An Order
    [Documentation]    AC-05.
    Ensure Purchase Order Line Under Test
    Should Not Be Empty    ${PURCHASE_ORDER_LINE_ID}

Line Totals Are Computed From Quantity And Price
    [Documentation]    subtotal = quantity x unit_price. 10 x 25.00 = 250.00.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00
    ${line}=    GET On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Response Status Should Be    ${line}    200
    Should Be Equal As Numbers    ${line.json()}[subtotal]    250.00
    Should Be Equal As Numbers    ${line.json()}[total]       250.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

Tax Is An Input And Not A Calculation
    [Documentation]    D9 and §55.15: there is no tax engine and the requirement forbids
    ...    building one. The client supplies a line's tax and the server sums it — which is why
    ...    tax_amount is writable on the line while subtotal and total are not.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00    12.50
    ${line}=    GET On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Should Be Equal As Numbers    ${line.json()}[tax_amount]    12.50
    Should Be Equal As Numbers    ${line.json()}[total]         262.50
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

A Discount Comes Off The Whole Line
    [Documentation]    The discount is applied to the line rather than the unit price, so a
    ...    fractional percentage is rounded once instead of once per unit. 4 x 25.00 less 10% is
    ...    90.00.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    4    25.00    0    10
    ${line}=    GET On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Should Be Equal As Numbers    ${line.json()}[subtotal]    90.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

A Client Supplied Subtotal Is Overwritten
    [Documentation]    §10.2. The client's number is ignored rather than rejected — echoing a
    ...    record back must not fail — but it must not be honoured either.
    Ensure Purchase Fixtures
    Ensure Purchase Product
    Ensure Purchase Uom
    ${order_id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_LINE_API}    json=${{ {'purchase_order_id': $order_id, 'sequence': 1, 'line_type': 'product', 'product_variant_id': $PURCHASE_VARIANT_ID, 'uom_id': $PURCHASE_UOM_ID, 'quantity': '2', 'unit_price': '10.00', 'discount_percent': '0', 'tax_amount': '0', 'subtotal': '999999.00', 'total': '999999.00', 'org_id': $PURCHASE_ORG_ID} }}
    ${line_id}    ${line_etag}=    Response Should Be Create Success    ${resp}
    ${line}=    GET On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Should Be Equal As Numbers    ${line.json()}[subtotal]    20.00
    Should Be Equal As Numbers    ${line.json()}[total]       20.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

The Header Totals Are The Sum Of The Lines
    [Documentation]    PUR-R4. Two lines, 250.00 and 100.00, and a header that agrees with the
    ...    column a reader can add up by hand.
    ${order_id}    ${etag}=    Create Purchase Order
    Create Purchase Order Line    ${order_id}    10    25.00    0    0    1
    Create Purchase Order Line    ${order_id}    4     25.00    0    0    2
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Response Status Should Be    ${order}    200
    Should Be Equal As Numbers    ${order.json()}[untaxed_amount]    350.00
    Should Be Equal As Numbers    ${order.json()}[total_amount]      350.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

The Header Follows A Line Update
    [Documentation]    The recompute runs on every line write, not only on create. A header that
    ...    tracked creates but not updates would drift the first time anybody corrected a
    ...    quantity.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}    json=${{ {'etag': $line_etag, 'quantity': '20'} }}
    Response Should Be Update Success    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Should Be Equal As Numbers    ${order.json()}[total_amount]    500.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

The Header Follows A Line Delete
    [Documentation]    Deleting the only line takes the order back to zero. The owning order is
    ...    read before the delete, because afterwards the line is gone and with it the only
    ...    pointer to the order whose totals just changed.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00
    ${resp}=    DELETE On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Response Should Be Delete Success    ${resp}
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Should Be Equal As Numbers    ${order.json()}[total_amount]    0
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

A Section Line Contributes Nothing To The Totals
    [Documentation]    A section organizes the printed order. Letting a heading carry money
    ...    would put a number in the total that no product accounts for.
    Ensure Purchase Fixtures
    ${order_id}    ${etag}=    Create Purchase Order
    Create Purchase Order Line    ${order_id}    10    25.00    0    0    1
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_LINE_API}    json=${{ {'purchase_order_id': $order_id, 'sequence': 2, 'line_type': 'section', 'description': 'Consumables', 'quantity': '0', 'unit_price': '0', 'discount_percent': '0', 'tax_amount': '0', 'org_id': $PURCHASE_ORG_ID} }}
    Response Status Should Be    ${resp}    201
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Should Be Equal As Numbers    ${order.json()}[total_amount]    250.00
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

The Ordered Quantity And Unit Survive The Conversion
    [Documentation]    AC-UOM-PUR-01 and BR-UOM-PUR-004, the central rule of PUR-R8. "10 boxes"
    ...    and "120 units" are the same goods but not the same request, and the purchase order
    ...    must say what was ordered. The converted figure lands in inventory_quantity instead
    ...    (003), which is what stock reads.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00
    ${line}=    GET On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}
    Response Status Should Be    ${line}    200
    Should Be Equal As Numbers    ${line.json()}[quantity]    10
    Should Be Equal    ${line.json()}[uom_id]    ${PURCHASE_UOM_ID}
    Dictionary Should Contain Key    ${line.json()}    inventory_quantity
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

Inventory Quantity Cannot Be Written By A Client
    [Documentation]    AC-UOM-PUR-03. It is derived from the ordered quantity through Essential's
    ...    conversion, so a client value would be a second answer to a question that already has
    ...    one — and the one Purchase stored would be the one nobody could reproduce.
    ${order_id}    ${etag}=    Create Purchase Order
    ${line_id}    ${line_etag}=    Create Purchase Order Line    ${order_id}    10    25.00
    ${resp}=    PUT On Session    api    ${PURCHASE_ORDER_LINE_API}/${line_id}    json=${{ {'etag': $line_etag, 'inventory_quantity': '99999'} }}    expected_status=any
    Response Should Be Client Error    ${resp}
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

A Line Refuses An Unknown Product
    [Documentation]    AC-UOM-PUR-05. A bad id is a different problem from a product the
    ...    business does not buy, and the two carry different violation keys so a buyer knows
    ...    which they hit.
    Ensure Purchase Fixtures
    Ensure Purchase Uom
    ${order_id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_LINE_API}    json=${{ {'purchase_order_id': $order_id, 'sequence': 1, 'line_type': 'product', 'product_variant_id': '01JZZZZZZZZZZZZZZZZZZZZZZZ', 'uom_id': $PURCHASE_UOM_ID, 'quantity': '1', 'unit_price': '1.00', 'discount_percent': '0', 'tax_amount': '0', 'org_id': $PURCHASE_ORG_ID} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order_line.product_not_found
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

A Line Refuses An Unknown Unit
    [Documentation]    AC-UOM-PUR-06. A typo would otherwise produce a line whose quantity is
    ...    expressed in nothing.
    Ensure Purchase Fixtures
    Ensure Purchase Product
    ${order_id}    ${etag}=    Create Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_LINE_API}    json=${{ {'purchase_order_id': $order_id, 'sequence': 1, 'line_type': 'product', 'product_variant_id': $PURCHASE_VARIANT_ID, 'uom_id': '01JZZZZZZZZZZZZZZZZZZZZZZZ', 'quantity': '1', 'unit_price': '1.00', 'discount_percent': '0', 'tax_amount': '0', 'org_id': $PURCHASE_ORG_ID} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order_line.uom_not_found
    [Teardown]    Delete Purchase Order Fixture    ${order_id}
