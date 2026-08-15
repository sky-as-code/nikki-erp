*** Settings ***
Documentation     The Inventory Location schema is served by the dynamic resource engine, same
...               as the other Inventory resources.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    inventory_location    schema


*** Test Cases ***
Get Inventory Location Model Schema
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    location_usage
    Dictionary Should Contain Key    ${fields}    parent_location_id
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares The Warehouse Fields
    [Documentation]    Location is the module's shared location master: Warehouse configures it
    ...    and Stock only references its id. These are the fields Warehouse configures.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    warehouse_id
    Dictionary Should Contain Key    ${fields}    purpose
    Dictionary Should Contain Key    ${fields}    complete_path
    Dictionary Should Contain Key    ${fields}    storage_category_id
    Dictionary Should Contain Key    ${fields}    removal_strategy
    Dictionary Should Contain Key    ${fields}    is_system_generated

Warehouse Is Optional On A Location
    [Documentation]    Not every location belongs to a warehouse. Vendor, customer,
    ...    inventory-loss and shared transit locations belong to none, and a transit location
    ...    between two warehouses must not belong to either.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${field}=    Set Variable    ${resp.json()}[fields][warehouse_id]
    Should Not Be True    ${field}[required_for_create]

Schema Declares Every Location Usage
    [Documentation]    BR §4.2: a movement always has two endpoints, so the counterparty and
    ...    virtual locations must exist alongside `internal`. Without `inventory_loss` an
    ...    adjustment has nowhere to balance against, and without `scrap` a scrap cannot move.
    ...    `scrap` stays distinct from `inventory_loss` for that reason.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][location_usage][data_type][options][enumValues]
    Should Contain    ${values}    internal
    Should Contain    ${values}    customer
    Should Contain    ${values}    vendor
    Should Contain    ${values}    inventory_loss
    Should Contain    ${values}    scrap
    Should Contain    ${values}    transit
    Should Not Contain    ${values}    supplier

Status Is Active Or Suspended Only
    [Documentation]    Status carries the operational state and is_archived carries withdrawal
    ...    from the working set; they are independent. There is deliberately no `inactive` and
    ...    no `blocked`: both were folded into `suspended`, and no `archived` either, because
    ...    archiving is not a status.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    Should Contain    ${values}    active
    Should Contain    ${values}    suspended
    Should Not Contain    ${values}    inactive
    Should Not Contain    ${values}    blocked
    Should Not Contain    ${values}    archived

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a location shows as an opaque id when chosen on a balance.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
