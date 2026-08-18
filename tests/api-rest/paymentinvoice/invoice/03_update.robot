*** Settings ***
Documentation     Updating Invoices. The success cases run first (they consume and rotate the
...               saved etag); negatives follow, and among them the system-managed fields a
...               client must never be able to set.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Invoice Under Test
Test Tags         paymentinvoice    invoice    update


*** Test Cases ***
Update Succeeds
    ${partner}=    Unique Display Name    Robot Updated Partner
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'partner_name': $partner, 'etag': $INVOICE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVOICE_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVOICE_ETAG}    ${etag}

Update Optional Fields Succeeds
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'note': 'Updated by the robot suite', 'partner_address': '2 Robot Street', 'etag': $INVOICE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVOICE_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVOICE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[note]    Updated by the robot suite
    Set Global Variable    ${INVOICE_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${partner}=    Unique Display Name    Robot Stale Partner
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'partner_name': $partner, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update Cannot Set The Status
    [Documentation]    status is no_update precisely so a draft cannot be declared issued by a
    ...    plain PATCH. Issuing assigns a number and freezes the totals; a client that could
    ...    skip that would produce an "issued" invoice with neither.
    ...
    ...    The engine answers 200 and silently drops the field rather than rejecting the
    ...    request — unlike number and the totals below, which are refused outright. What
    ...    matters either way is the stored value, so that is what is asserted: a test pinned
    ...    to the status code would pass while the status changed underneath it.
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'status': 'issued', 'etag': $INVOICE_ETAG} }}    expected_status=any
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[status]    draft
    ...    msg=A client must not be able to issue an invoice through a plain update
    #    The etag rotated if the PATCH was accepted, so it is re-read for the tests that follow.
    Set Global Variable    ${INVOICE_ETAG}    ${item}[etag]

Update Cannot Set The Number
    [Documentation]    The number comes from the issue action's sequence. A client-chosen one
    ...    would collide with, or leave a gap in, that sequence. As with the status, the
    ...    stored value is what is asserted rather than the status code.
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'number': 'INV-2026-000001', 'etag': $INVOICE_ETAG} }}    expected_status=any
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    ${number}=    Get From Dictionary    ${item}    number    ${None}
    Should Be Equal    ${number}    ${None}
    ...    msg=A client must not be able to choose an invoice number
    Set Global Variable    ${INVOICE_ETAG}    ${item}[etag]

Update Cannot Set The Totals
    [Documentation]    The totals are recomputed from the lines on issue. A client that could
    ...    write them directly could author a document whose total disagrees with what it
    ...    totals, which is the one thing an invoice must never do.
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    json=${{ {'total_amount': '999999', 'etag': $INVOICE_ETAG} }}    expected_status=any
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal As Numbers    ${item}[total_amount]    0
    ...    msg=A client must not be able to write an invoice total directly
    Set Global Variable    ${INVOICE_ETAG}    ${item}[etag]

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${INVOICE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'note': 'x', 'etag': $INVOICE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
