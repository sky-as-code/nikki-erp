*** Settings ***
Documentation     The Warehouse schema is served by the dynamic resource engine, same as the
...               other Inventory resources.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    warehouse    schema


*** Test Cases ***
Get Warehouse Model Schema
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    warehouse_role
    Dictionary Should Contain Key    ${fields}    parent_warehouse_id
    Dictionary Should Contain Key    ${fields}    incoming_flow
    Dictionary Should Contain Key    ${fields}    outgoing_flow
    Dictionary Should Contain Key    ${fields}    status
    Dictionary Should Contain Key    ${fields}    org_id

Status Is Active Or Suspended Only
    [Documentation]    Status carries the operational state and is_archived carries withdrawal
    ...    from the working set; the two are independent. There is deliberately no `inactive`,
    ...    which was folded into `suspended`, and no `archived`, because archiving is not a
    ...    status.
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][status][data_type][options][enumValues]
    Should Contain    ${values}    active
    Should Contain    ${values}    suspended
    Should Not Contain    ${values}    inactive
    Should Not Contain    ${values}    archived

Schema Declares Every Flow Setting
    [Documentation]    One, two or three stops in each direction. A fourth value would have no
    ...    topology behind it: the location provisioning and the movement plan are both
    ...    switches over exactly these three.
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    FOR    ${field}    IN    incoming_flow    outgoing_flow
        ${values}=    Set Variable    ${resp.json()}[fields][${field}][data_type][options][enumValues]
        Should Contain    ${values}    one_step
        Should Contain    ${values}    two_step
        Should Contain    ${values}    three_step
    END

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a warehouse shows as an opaque id when chosen on a location.
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
