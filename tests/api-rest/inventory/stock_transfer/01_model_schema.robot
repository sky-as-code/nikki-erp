*** Settings ***
Documentation     The Stock Transfer schema as the engine advertises it at meta/schema.
...
...               The fields asserted here are the ones a client cannot be allowed to set by
...               hand, and the ones the movement engine reads to decide what to do. A schema
...               that stops advertising them still serves CRUD, so nothing else in the tree
...               would notice they had gone.
Library           Collections
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    stock_transfer    schema


*** Test Cases ***
Schema Declares The Transaction Fields
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[data][fields]
    ${names}=    Evaluate    [f['name'] for f in $fields]
    FOR    ${field}    IN
    ...    transfer_number    operation_type_id    operation_code    status
    ...    source_location_id    destination_location_id
        Should Contain    ${names}    ${field}
        ...    msg=The transfer schema must declare ${field}
    END

Schema Declares The Snapshot Policies
    [Documentation]    BR §4.2.3.4: a transfer carries its own copy of the operation type's
    ...    policies, so that reconfiguring the type cannot reinterpret a transfer already
    ...    created. Reading them through the edge instead would lose that guarantee.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/meta/schema
    ${fields}=    Set Variable    ${resp.json()}[data][fields]
    ${names}=    Evaluate    [f['name'] for f in $fields]
    FOR    ${field}    IN    reservation_method    backorder_policy    shipping_policy
        Should Contain    ${names}    ${field}
        ...    msg=${field} must be snapshotted onto the transfer, not read through the operation type
    END

Schema Declares The Backorder And Idempotency Links
    [Documentation]    backorder_of_id carries STOCK-INV-010 and idempotency_key carries BR §8.7.
    ...    Both are what make a partial delivery and a retried validate traceable after the fact.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/meta/schema
    ${fields}=    Set Variable    ${resp.json()}[data][fields]
    ${names}=    Evaluate    [f['name'] for f in $fields]
    Should Contain    ${names}    backorder_of_id
    Should Contain    ${names}    idempotency_key
    Should Contain    ${names}    completed_at

Status Is Not Client Updatable
    [Documentation]    AC-STOCK-002 in spirit: a status set by hand would claim an outcome that
    ...    no movement produced. It is advanced only by confirm, reserve, validate and cancel.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/meta/schema
    ${fields}=    Set Variable    ${resp.json()}[data][fields]
    ${status}=    Evaluate    next(f for f in $fields if f['name'] == 'status')
    Should Be True    ${status}[no_update]
    ...    msg=status must be no_update: it is derived from the transfer's moves, never assigned
