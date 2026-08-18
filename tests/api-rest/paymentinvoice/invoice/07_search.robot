*** Settings ***
Documentation     Searching Invoices. An invoice is org-scoped, so every search carries
...               org_id.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Invoice Under Test
Test Tags         paymentinvoice    invoice    search


*** Variables ***
${INVOICE_SCHEMA}    ${PAYINV_SCHEMA_DIR}/invoice.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${INVOICE_API}    params=${{ {'org_id': $PAYINV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${INVOICE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${INVOICE_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'page': 0, 'size': 5} }}
    Response Should Be Search Success    ${resp}    ${INVOICE_SCHEMA}    size=5    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${INVOICE_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${INVOICE_SCHEMA}    size=50    page=99    item_count=0

Search By Status Succeeds
    [Documentation]    Filtering by status is what the frontend's draft/issued tabs do.
    ${resp}=    GET On Session    api    ${INVOICE_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'graph': '{"if":["status", "=", "draft"]}'} }}
    Response Should Be Search Success    ${resp}    ${INVOICE_SCHEMA}    size=50    page=0

Search Finds The Invoice Under Test
    ${resp}=    GET On Session    api    ${INVOICE_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'graph': '{"if":["id", "=", "' + $INVOICE_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${INVOICE_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${INVOICE_ID}

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${INVOICE_API}
    ...    params=${{ {'org_id': $PAYINV_ORG_ID, 'graph': '{"if":["bla_bla_field", "=", "x"]}'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
