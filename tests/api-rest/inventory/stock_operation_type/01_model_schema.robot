*** Settings ***
Documentation     The Stock Operation Type schema is served by the dynamic resource engine,
...               same as the other Inventory resources.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    stock_operation_type    schema


*** Test Cases ***
Get Stock Operation Type Model Schema
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    operation_code
    Dictionary Should Contain Key    ${fields}    reservation_method
    Dictionary Should Contain Key    ${fields}    backorder_policy
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares The Policy Enums
    [Documentation]    BR §4.2.1.2 and §4.2.6.4: reservation method and backorder policy are
    ...    what a transfer snapshots at creation, so the accepted values are pinned here.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    ${operations}=    Set Variable    ${fields}[operation_code][data_type][options][enumValues]
    Should Contain    ${operations}    incoming
    Should Contain    ${operations}    outgoing
    Should Contain    ${operations}    internal
    ${methods}=    Set Variable    ${fields}[reservation_method][data_type][options][enumValues]
    Should Contain    ${methods}    at_confirmation
    Should Contain    ${methods}    manual
    Should Contain    ${methods}    before_scheduled_date
    ${policies}=    Set Variable    ${fields}[backorder_policy][data_type][options][enumValues]
    Should Contain    ${policies}    ask
    Should Contain    ${policies}    always
    Should Contain    ${policies}    never

Schema Declares No Status Field
    [Documentation]    BR §4.2.1.2 states it plainly: an operation type has no business
    ...    lifecycle, only archive. Adding a status would give "retired" two meanings.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Marks Operation Code Immutable
    [Documentation]    A transfer snapshots the direction of its type. Letting the type flip
    ...    from incoming to outgoing would reinterpret movements already recorded against it.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[fields][operation_code][no_update]    ${True}
