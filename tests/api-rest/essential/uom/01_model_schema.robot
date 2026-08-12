*** Settings ***
Documentation     The UoM schema drives the frontend form, so the field set and the
...               decimal typing of factor/rounding are contracts, not incidentals.
Resource          resources/essential.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    uom    schema


*** Test Cases ***
Get Uom Model Schema
    ${resp}=    GET On Session    api    ${UOM_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Factor And Rounding Are Decimal
    [Documentation]    BR-UOM-ESS-018: the conversion factor must hold values such as
    ...    0.453592 without float drift. An integer type here would silently truncate
    ...    every sub-unit factor.
    ${resp}=    GET On Session    api    ${UOM_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Should Be Equal    ${fields}[factor][data_type][name]    decimal
    Should Be Equal    ${fields}[rounding][data_type][name]    decimal

Uom Type Is A Closed Enum
    [Documentation]    BR-UOM-ESS-009: the three types are the closed set the validation
    ...    rules switch on, and the frontend renders them as a Select from this list.
    ${resp}=    GET On Session    api    ${UOM_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][uom_type][data_type][options][enumValues]
    Lists Should Be Equal    ${values}    ${{ ['reference', 'bigger_equal', 'smaller'] }}
    ...    ignore_order=True

Category Is Declared As A Relation
    [Documentation]    The category picker on the UoM form resolves through to_relations;
    ...    without the edge the field degrades to a raw ULID text input.
    ${resp}=    GET On Session    api    ${UOM_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${relations}=    Set Variable    ${resp.json()}[to_relations]
    ${src_fields}=    Evaluate    [rel.get('src_field') for rel in $relations]
    Should Contain    ${src_fields}    category_id
