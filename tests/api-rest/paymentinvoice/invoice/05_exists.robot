*** Settings ***
Documentation     Existence checks over Invoices.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Invoice Under Test
Test Tags         paymentinvoice    invoice    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${INVOICE_API}/exists
    ...    json=${{ {'ids': [$INVOICE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists Separates Present From Absent
    ${fakes}=    Not Found Id List    2
    ${ids}=    Combine Lists    ${{ [$INVOICE_ID] }}    ${fakes}
    ${resp}=    POST On Session    api    ${INVOICE_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=2

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVOICE_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVOICE_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
