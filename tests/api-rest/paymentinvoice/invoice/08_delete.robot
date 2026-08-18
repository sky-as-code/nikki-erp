*** Settings ***
Documentation     Deleting the Invoice under test — always the LAST suite, doubling as
...               cleanup. The line goes first: it points at the invoice, and that FK is
...               ON DELETE CASCADE in the generated schema, so deleting the line explicitly
...               is what proves the line endpoint works rather than relying on the cascade.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Invoice Under Test
Test Tags         paymentinvoice    invoice    delete


*** Test Cases ***
Delete Invoice Line Succeeds
    ${line_id}=    Get Variable Value    ${INVOICE_LINE_ID}    ${EMPTY}
    Skip If    not $line_id    No invoice line was created by this run
    ${resp}=    DELETE On Session    api    ${INVOICE_LINE_API}/${line_id}
    Response Should Be Delete Success    ${resp}    count=1
    Set Global Variable    ${INVOICE_LINE_ID}    ${EMPTY}

Delete Succeeds
    ${resp}=    DELETE On Session    api    ${INVOICE_API}/${INVOICE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    invoice under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVOICE_API}/${INVOICE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${INVOICE_ID}    ${EMPTY}
    Set Global Variable    ${INVOICE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVOICE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVOICE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
