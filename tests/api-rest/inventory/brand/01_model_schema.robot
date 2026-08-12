*** Settings ***
Documentation     The Brand schema is served by the dynamic resource engine, same as the
...               other six Inventory resources.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    brand    schema


*** Test Cases ***
Get Brand Model Schema
    ${resp}=    GET On Session    api    ${BRAND_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${BRAND_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    code
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    website
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Status Field
    [Documentation]    BR §6.8.1: a brand's lifecycle is archive alone. A status field would
    ...    invite a second, conflicting notion of "retired".
    ${resp}=    GET On Session    api    ${BRAND_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a brand shows as an opaque id when chosen from a template.
    ${resp}=    GET On Session    api    ${BRAND_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
