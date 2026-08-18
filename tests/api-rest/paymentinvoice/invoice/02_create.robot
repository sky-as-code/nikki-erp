*** Settings ***
Documentation     Creating Invoices. The first test saves the invoice under test
...               (${INVOICE_ID}/${INVOICE_ETAG}) consumed by the later suites and deleted
...               last by 08_delete.robot.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Payment Invoice Org
...               AND    Ensure Payment Currency
Test Tags         paymentinvoice    invoice    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${partner}=    Unique Display Name    Robot Partner
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    json=${{ {'partner_name': $partner, 'currency_id': $PAYINV_CURRENCY_ID, 'org_id': $PAYINV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INVOICE_ID}    ${id}
    Set Global Variable    ${INVOICE_ETAG}    ${etag}

A New Invoice Is A Draft With Zero Totals
    [Documentation]    status defaults to draft and the three totals to zero, because an
    ...    invoice's figures are computed from its lines on issue and are meaningless before
    ...    then. A client that could set them on create could author a document whose total
    ...    disagreed with what it totals.
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[status]    draft
    Should Be Equal As Numbers    ${item}[subtotal_amount]    0
    Should Be Equal As Numbers    ${item}[tax_amount]    0
    Should Be Equal As Numbers    ${item}[total_amount]    0

A New Invoice Has No Number
    [Documentation]    The number is assigned by the issue action, not on create: handing one
    ...    out earlier would leave a gap in the sequence whenever a draft was abandoned.
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    ${number}=    Get From Dictionary    ${item}    number    ${None}
    Should Be Equal    ${number}    ${None}
    ...    msg=A draft invoice must not carry a number before it is issued

Create With All Optional Fields Succeeds
    ${partner}=    Unique Display Name    Robot Full Partner
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    json=${{ {'partner_name': $partner, 'currency_id': $PAYINV_CURRENCY_ID, 'org_id': $PAYINV_ORG_ID, 'partner_tax_code': '0101234567', 'partner_address': '1 Robot Street', 'note': 'Created by the robot suite'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${INVOICE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[partner_tax_code]    0101234567
    Should Be Equal    ${item}[partner_address]    1 Robot Street
    DELETE On Session    api    ${INVOICE_API}/${id}    expected_status=any

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVOICE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    partner_name    currency_id    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    data={ "partner_name": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${partner}=    Unique Display Name    Robot Nonexist Field
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    json=${{ {'partner_name': $partner, 'currency_id': $PAYINV_CURRENCY_ID, 'org_id': $PAYINV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Create Invoice Line Succeeds
    [Documentation]    Saves the line under test, which 06_issue.robot totals and
    ...    08_delete.robot removes.
    ${line_id}=    Add Invoice Line    ${INVOICE_ID}    2    50000    0
    Set Global Variable    ${INVOICE_LINE_ID}    ${line_id}
    ${resp}=    GET On Session    api    ${INVOICE_LINE_API}/${line_id}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice_line.json    200
    Should Be Equal As Integers    ${item}[quantity]    2

Create Invoice Line With Missing Required Fields Fails
    [Documentation]    Every field without a default must be reported, not the first one: a
    ...    caller fixing them one round trip at a time is the alternative. quantity,
    ...    tax_rate_percent and amount are absent from this list although they are
    ...    required-for-create — each carries a default_value, so the engine fills it rather
    ...    than reporting it missing.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVOICE_LINE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}
    ...    invoice_id    description    unit_price    org_id
