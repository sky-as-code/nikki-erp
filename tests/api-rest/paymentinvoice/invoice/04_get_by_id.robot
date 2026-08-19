*** Settings ***
Documentation     Reading one Invoice back.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Invoice Under Test
Test Tags         paymentinvoice    invoice    get


*** Test Cases ***
Get By Id Succeeds
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/invoice.json    200
    Should Be Equal    ${item}[id]    ${INVOICE_ID}

Get By Id With Selected Fields Succeeds
    [Documentation]    A projection returns only the named columns, which is what keeps a
    ...    listing from fetching every field of every row.
    ${resp}=    GET On Session    api    ${INVOICE_API}/${INVOICE_ID}
    ...    params=${{ {'fields': 'id,status,partner_name'} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[item]
    Dictionary Should Contain Key    ${item}    status
    Dictionary Should Contain Key    ${item}    partner_name
    Dictionary Should Not Contain Key    ${item}    note

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${INVOICE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${INVOICE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
