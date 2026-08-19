*** Settings ***
Documentation     The Invoice schema is served by the dynamic resource engine, same as every
...               other resource in this module.
Resource          resources/paymentinvoice.resource
Suite Setup       Create Authorized API Session
Test Tags         paymentinvoice    invoice    schema


*** Test Cases ***
Get Invoice Model Schema
    ${resp}=    GET On Session    api    ${INVOICE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${INVOICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    number
    Dictionary Should Contain Key    ${fields}    status
    Dictionary Should Contain Key    ${fields}    partner_name
    Dictionary Should Contain Key    ${fields}    currency_id
    Dictionary Should Contain Key    ${fields}    subtotal_amount
    Dictionary Should Contain Key    ${fields}    tax_amount
    Dictionary Should Contain Key    ${fields}    total_amount
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Gateway Specific Column
    [Documentation]    An invoice accounts for a sale and knows nothing about how it was
    ...    collected. A gateway-specific column here would make the accounting document
    ...    depend on which payment provider happened to serve it.
    ${resp}=    GET On Session    api    ${INVOICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Not Contain Key    ${fields}    pos_id
    Dictionary Should Not Contain Key    ${fields}    momo_trans_id

Schema Declares Every Invoice Status
    [Documentation]    The four statuses are the invoice's whole lifecycle. A value missing
    ...    here is one the frontend cannot render and the issue action cannot reach.
    ${resp}=    GET On Session    api    ${INVOICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    Should Contain    ${values}    draft
    Should Contain    ${values}    issued
    Should Contain    ${values}    paid
    Should Contain    ${values}    void

Get Invoice Line Model Schema
    ${resp}=    GET On Session    api    ${INVOICE_LINE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Line Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${INVOICE_LINE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    invoice_id
    Dictionary Should Contain Key    ${fields}    quantity
    Dictionary Should Contain Key    ${fields}    unit_price
    Dictionary Should Contain Key    ${fields}    tax_rate_percent
    Dictionary Should Contain Key    ${fields}    amount
