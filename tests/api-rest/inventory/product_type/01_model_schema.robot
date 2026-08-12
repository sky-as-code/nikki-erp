*** Settings ***
Documentation     The Product Type schema is served by the dynamic resource engine, so this
...               also proves Inventory's engine routes are registered at all — this module
...               had no engine-served resources before PROD-002.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_type    schema


*** Test Cases ***
Get Product Type Model Schema
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Capability Flags
    [Documentation]    BR §6.3.2: the four supports_* flags are what processing logic keys
    ...    off. The frontend renders the form from this schema, so a missing field is
    ...    invisible rather than a loud failure.
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    supports_stock
    Dictionary Should Contain Key    ${fields}    supports_sale
    Dictionary Should Contain Key    ${fields}    supports_purchase
    Dictionary Should Contain Key    ${fields}    supports_manufacturing

Schema Declares No Status Field
    [Documentation]    BR §6.3.2: a product type's lifecycle is archive alone. A status
    ...    field would invite a second, conflicting notion of "retired".
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a type shows as an opaque id when chosen from a template.
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
