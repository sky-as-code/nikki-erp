*** Settings ***
Documentation     The Product Attribute Value schema is served by the dynamic resource engine.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_attribute_value    schema


*** Test Cases ***
Get Attribute Value Model Schema
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Value Fields
    [Documentation]    A value belongs to an attribute (attribute_id), carries its own code/
    ...    name and an optional signed price_extra, and is owned by an org.
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    attribute_id
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    price_extra
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Status Field
    [Documentation]    An attribute value has no lifecycle beyond archive; a status field
    ...    would invite a second, conflicting notion of "retired".
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a value shows as an opaque id when chosen for a variant.
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
