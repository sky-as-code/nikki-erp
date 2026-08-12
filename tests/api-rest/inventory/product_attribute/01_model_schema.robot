*** Settings ***
Documentation     The Product Attribute schema is served by the dynamic resource engine.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_attribute    schema


*** Test Cases ***
Get Product Attribute Model Schema
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Variant Generation Fields
    [Documentation]    BR §6.5.3 / §14.3 step 2: data_type and variant_creation_mode are the
    ...    two enums that drive variant generation, so they must be visible on the schema
    ...    alongside the other create-required fields.
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    data_type
    Dictionary Should Contain Key    ${fields}    variant_creation_mode
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Status Field
    [Documentation]    Like every other Inventory master-data resource, a product attribute's
    ...    lifecycle is archive alone. A status field would invite a second, conflicting
    ...    notion of "retired".
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so an attribute shows as an opaque id when chosen from a template.
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
