*** Settings ***
Documentation     The Product Category schema is served by the dynamic resource engine.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_category    schema


*** Test Cases ***
Get Product Category Model Schema
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Tree Fields
    [Documentation]    BR §6.4: code/name identify the category, parent_category_id is the
    ...    self-FK the tree rule walks, org_id scopes it. A missing field here is invisible
    ...    on the frontend rather than a loud failure.
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    parent_category_id
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Status Field
    [Documentation]    BR §6.4.2: a category's lifecycle is archive alone. A status field
    ...    would invite a second, conflicting notion of "retired".
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a category shows as an opaque id when chosen as a parent.
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
