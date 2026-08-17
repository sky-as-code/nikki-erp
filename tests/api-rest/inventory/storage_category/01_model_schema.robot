*** Settings ***
Documentation     The Storage Category schema. It has no status field, which is the point: a
...               category is either part of the master data available for new assignments or
...               it is archived, and a second flag would say the same thing twice.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    storage_category    schema


*** Test Cases ***
Get Storage Category Model Schema
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    max_weight
    Dictionary Should Contain Key    ${fields}    allow_new_item_policy

Schema Declares No Status Field
    [Documentation]    Only Warehouse and Inventory Location have an operational state that is
    ...    independent of archiving. A category has no such state, so archiving is its whole
    ...    lifecycle and there is no suspend or resume to pair with a status.
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Does Not Store Used Capacity
    [Documentation]    What is currently stored depends on actual inventory, which Stock owns.
    ...    A number cached on master data goes stale the moment a move commits.
    ${resp}=    GET On Session    api    ${STORAGE_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Not Contain Key    ${fields}    used_capacity
    Dictionary Should Not Contain Key    ${fields}    available_capacity
