*** Settings ***
Documentation     The Order and Transaction schemas, served by the dynamic resource engine.
Resource          resources/paymentinvoice.resource
Suite Setup       Create Authorized API Session
Test Tags         paymentinvoice    order    schema


*** Test Cases ***
Get Order Model Schema
    ${resp}=    GET On Session    api    ${ORDER_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    order_id
    Dictionary Should Contain Key    ${fields}    order_code
    Dictionary Should Contain Key    ${fields}    status
    Dictionary Should Contain Key    ${fields}    amount
    Dictionary Should Contain Key    ${fields}    refund_amount
    Dictionary Should Contain Key    ${fields}    currency_id
    Dictionary Should Contain Key    ${fields}    payment_method_id
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares Every Order Status
    [Documentation]    A status missing here is one the state machine can reach and the
    ...    frontend cannot render.
    ${resp}=    GET On Session    api    ${ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    Should Contain    ${values}    pending
    Should Contain    ${values}    processing
    Should Contain    ${values}    payment_success
    Should Contain    ${values}    payment_failed
    Should Contain    ${values}    canceled
    Should Contain    ${values}    refund_success
    Should Contain    ${values}    refund_failed
    Should Contain    ${values}    expired

Schema Declares No Gateway Specific Column
    [Documentation]    Whatever a gateway needs at order time lives in the metadata map, which
    ...    each adapter owns the keys of. A column named after one gateway would make the
    ...    order table grow every time a gateway is added.
    ${resp}=    GET On Session    api    ${ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    metadata
    Dictionary Should Not Contain Key    ${fields}    pos_id
    Dictionary Should Not Contain Key    ${fields}    momo_trans_id
    Dictionary Should Not Contain Key    ${fields}    vietqr_reference_number

Schema Declares The Sync Bookkeeping
    [Documentation]    last_sync_status is what the retry sweep filters on and sync_logs is
    ...    what a human reads when a caller says it was never told a payment settled. The
    ...    service this module replaces wrote both and read neither.
    ${resp}=    GET On Session    api    ${ORDER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    last_sync_status
    Dictionary Should Contain Key    ${fields}    sync_logs

Get Transaction Model Schema
    ${resp}=    GET On Session    api    ${TRANSACTION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Transaction Schema Declares Both Directions Of Money
    [Documentation]    One payment transaction is written with the order and one refund
    ...    transaction is appended per successful refund, so both types must exist.
    ${resp}=    GET On Session    api    ${TRANSACTION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][transaction_type][data_type][options][enumValues]
    Should Contain    ${values}    payment
    Should Contain    ${values}    refund

Get Payment Method Model Schema
    ${resp}=    GET On Session    api    ${PAYMENT_METHOD_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Payment Method Schema Names The Adapter That Serves It
    [Documentation]    adapter_code is what makes offering a gateway under a second merchant
    ...    account a row rather than a release.
    ${resp}=    GET On Session    api    ${PAYMENT_METHOD_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}[fields]    adapter_code
