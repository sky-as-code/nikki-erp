*** Settings ***
Documentation     The UoM Category schema is served by the dynamic resource engine, so
...               this also proves the engine's routes are registered at all.
Resource          resources/essential.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    uomcat    schema


*** Test Cases ***
Get Uom Category Model Schema
    ${resp}=    GET On Session    api    ${UOMCAT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Reference Uom Field
    [Documentation]    BR-UOM-ESS-003: a category carries its single Reference UoM. The
    ...    frontend renders the form from this schema, so a missing field is invisible
    ...    rather than a loud failure.
    ${resp}=    GET On Session    api    ${UOMCAT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    reference_uom_id
    Dictionary Should Contain Key    ${fields}    name
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back
    ...    to the raw ULID, so a category shows as an opaque id when chosen from a UoM.
    ${resp}=    GET On Session    api    ${UOMCAT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
