*** Settings ***
Documentation     The Stock Location schema is served by the dynamic resource engine, same
...               as the other Inventory resources.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    stock_location    schema


*** Test Cases ***
Get Stock Location Model Schema
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    location_type
    Dictionary Should Contain Key    ${fields}    parent_location_id
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares Every Location Type
    [Documentation]    BR §4.2: a movement always has two endpoints, so the counterparty and
    ...    virtual locations must exist alongside `internal`. Without `inventory_loss` an
    ...    adjustment has nowhere to balance against, and without `scrap` a scrap cannot move.
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][location_type][data_type][options][enumValues]
    Should Contain    ${values}    internal
    Should Contain    ${values}    customer
    Should Contain    ${values}    supplier
    Should Contain    ${values}    inventory_loss
    Should Contain    ${values}    scrap
    Should Contain    ${values}    transit

Schema Declares No Status Field
    [Documentation]    A location's lifecycle is archive alone. A status field would invite a
    ...    second, conflicting notion of "retired". See BR §3.2.
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a location shows as an opaque id when chosen on a balance.
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
