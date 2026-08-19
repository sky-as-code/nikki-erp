*** Settings ***
Documentation     Issuing Invoices — the action that turns a draft into an accounting
...               document. It takes the archive slot the other resources use, and runs after
...               the read suites because it is irreversible: an issued invoice cannot go back
...               to draft.
...
...               Each test that needs an issued invoice mints its own draft rather than
...               sharing one, precisely because issuing is one-way.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Payment Invoice Org
...               AND    Ensure Payment Currency
Test Tags         paymentinvoice    invoice    issue


*** Test Cases ***
Issue Assigns A Number And Freezes The Totals
    [Documentation]    2 x 50000 with no tax is 100000. The totals are computed from the
    ...    lines rather than read from the draft, which is what makes an issued invoice agree
    ...    with what it totals.
    ${invoice_id}=    Create Draft Invoice
    Add Invoice Line    ${invoice_id}    2    50000    0
    ${resp}=    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal As Numbers    ${body}[subtotal_amount]    100000
    Should Be Equal As Numbers    ${body}[tax_amount]    0
    Should Be Equal As Numbers    ${body}[total_amount]    100000
    Should Match Regexp    ${body}[number]    ^INV-\\d{4}-\\d{6,}$

An Issued Invoice Is No Longer A Draft
    ${invoice_id}=    Create Draft Invoice
    Add Invoice Line    ${invoice_id}    1    25000    0
    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    ${resp}=    GET On Session    api    ${INVOICE_API}/${invoice_id}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[status]    issued
    Should Not Be Equal    ${item}[issued_at]    ${None}
    ...    msg=An issued invoice must carry the date it was issued

Issue Applies Tax Per Line
    [Documentation]    1 x 100000 at 10% is 10000 of tax on a 100000 subtotal. Tax accumulates
    ...    per line rather than being applied to the subtotal, because lines may carry
    ...    different rates.
    ${invoice_id}=    Create Draft Invoice
    Add Invoice Line    ${invoice_id}    1    100000    10
    ${resp}=    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal As Numbers    ${body}[subtotal_amount]    100000
    Should Be Equal As Numbers    ${body}[tax_amount]    10000
    Should Be Equal As Numbers    ${body}[total_amount]    110000

Issue Totals Lines Carrying Different Tax Rates
    [Documentation]    The case a single rate over the subtotal would get wrong: 100000 at 10%
    ...    plus 200000 at 0% is 10000 of tax, not 30000.
    ${invoice_id}=    Create Draft Invoice
    Add Invoice Line    ${invoice_id}    1    100000    10
    Add Invoice Line    ${invoice_id}    1    200000    0
    ${resp}=    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal As Numbers    ${body}[subtotal_amount]    300000
    Should Be Equal As Numbers    ${body}[tax_amount]    10000
    Should Be Equal As Numbers    ${body}[total_amount]    310000

Issue Recomputes A Line Amount That Disagrees With Its Quantity And Price
    [Documentation]    The line is created with amount 0 while quantity x unit_price is 60000.
    ...    The quantity and the price are what a reader can verify, so they win and the stored
    ...    amount is corrected rather than trusted.
    ${invoice_id}=    Create Draft Invoice
    ${line_id}=    Add Invoice Line    ${invoice_id}    3    20000    0
    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    ${resp}=    GET On Session    api    ${INVOICE_LINE_API}/${line_id}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice_line.json    200
    Should Be Equal As Numbers    ${item}[amount]    60000

Two Invoices Issued In Sequence Get Distinct Numbers
    [Documentation]    Two documents sharing a number is the defect an auditor finds. The
    ...    sequence is derived inside the issue transaction, so consecutive issues advance it.
    ${first_id}=    Create Draft Invoice
    Add Invoice Line    ${first_id}    1    10000    0
    ${second_id}=    Create Draft Invoice
    Add Invoice Line    ${second_id}    1    10000    0
    ${resp}=    POST On Session    api    ${INVOICE_API}/${first_id}/issue    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${first_number}=    Set Variable    ${resp.json()}[number]
    ${resp}=    POST On Session    api    ${INVOICE_API}/${second_id}/issue    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${second_number}=    Set Variable    ${resp.json()}[number]
    Should Not Be Equal    ${first_number}    ${second_number}
    ...    msg=Two invoices must never be issued under one number

Issuing An Already Issued Invoice Fails
    [Documentation]    Re-issuing would mint a second number for one document and re-freeze
    ...    totals that may since have been paid against.
    [Tags]    negative
    ${invoice_id}=    Create Draft Invoice
    Add Invoice Line    ${invoice_id}    1    10000    0
    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue    json=${{ {} }}
    ${resp}=    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue
    ...    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    ...    msg=A second issue must be refused as a client error, not a server failure

An Invoice With No Lines Cannot Be Issued
    [Documentation]    It would total zero, which is not a document anyone meant to issue.
    [Tags]    negative
    ${invoice_id}=    Create Draft Invoice
    ${resp}=    POST On Session    api    ${INVOICE_API}/${invoice_id}/issue
    ...    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    ...    msg=An invoice with no lines must be refused
    ${resp}=    GET On Session    api    ${INVOICE_API}/${invoice_id}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[status]    draft
    ...    msg=A refused issue must leave the invoice a draft

Issuing A Nonexistent Invoice Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVOICE_API}/${NOT_FOUND_ID}/issue
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Issuing an invoice that does not exist must not succeed
