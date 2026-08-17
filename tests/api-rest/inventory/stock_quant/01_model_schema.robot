*** Settings ***
Documentation     The Stock Quant schema is served by the dynamic resource engine, same as
...               the other Inventory resources, but its shape encodes the balance rules:
...               no status, no is_archived, and a derived available quantity.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    stock_quant    schema


*** Test Cases ***
Get Stock Quant Model Schema
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Balance Dimensions
    [Documentation]    BR §4.2.2.2: a balance is identified by product, location, lot,
    ...    package and owner together.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    product_variant_id
    Dictionary Should Contain Key    ${fields}    location_id
    Dictionary Should Contain Key    ${fields}    lot_ref
    Dictionary Should Contain Key    ${fields}    package_ref
    Dictionary Should Contain Key    ${fields}    owner_ref
    Dictionary Should Contain Key    ${fields}    on_hand_quantity
    Dictionary Should Contain Key    ${fields}    reserved_quantity

Schema Declares Available Quantity As Virtual
    [Documentation]    BR §4.2.2.3: available is on-hand minus reserved, computed on read.
    ...    Storing it would be a third number that can disagree with the two it derives from.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${field}=    Set Variable    ${resp.json()}[fields][available_quantity]
    Should Be Equal    ${field}[is_virtual]    ${True}
    Should Be Equal    ${field}[is_computed]    ${True}
    Should Be Equal    ${field}[is_persisted]    ${False}
    Should Be Equal    ${field}[is_edge_model]    ${False}
    # Read-only but not server-owned: a computed field carries business meaning, so it stays
    # selectable as a column rather than being filtered out with the keys.
    Should Be Equal    ${field}[is_system_field]    ${False}

Schema Declares No Status Or Archive Field
    [Documentation]    BR §4.2.2.2: a quant is current state, so it has neither a business
    ...    lifecycle nor an archive flag. A "retired balance" is not a concept.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Not Contain Key    ${fields}    status
    Dictionary Should Not Contain Key    ${fields}    is_archived

Schema Marks The Quantities Not Updatable
    [Documentation]    AC-STOCK-002: no business API may write on-hand directly. Marking the
    ...    columns no_update is the schema-level half of that; 03_reject_direct_write.robot
    ...    pins the engine-level half.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Should Be Equal    ${fields}[on_hand_quantity][no_update]    ${True}
    Should Be Equal    ${fields}[reserved_quantity][no_update]    ${True}

Schema Declares The Count Fields
    [Documentation]    BR §4.2.7 and §4.2.8: a physical count is recorded on the quant, and
    ...    the snapshot is what lets a stale count be rejected at apply time.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    counted_quantity
    Dictionary Should Contain Key    ${fields}    count_quantity_set
    Dictionary Should Contain Key    ${fields}    count_snapshot_quantity
    Dictionary Should Contain Key    ${fields}    next_count_date
    Dictionary Should Contain Key    ${fields}    last_count_date
    Dictionary Should Contain Key    ${fields}    count_assigned_user_id
