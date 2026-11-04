*** Settings ***
Documentation     The definition schema drives the admin form; the routes it serves are part of
...               the contract too: no delete (BR §4.4) and no way to issue a number (BR §6.1).
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    sequence_definition    schema


*** Test Cases ***
Get Sequence Definition Model Schema
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Configuration Fields
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    configuration_scope    org_id    owner_module    code    name    sub_scope
    ...    scope_type    format_owner    algorithm    prefix_template    suffix_template    padding
    ...    increment    initial_value    reset_period    timezone    issuance_retention_days
    ...    first_issued_at    is_archived    etag
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Reset Period Is A Closed Enum
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][reset_period][data_type][options][enumValues]
    Lists Should Be Equal    ${values}    ${{ ['never', 'yearly', 'monthly', 'daily'] }}    ignore_order=True

Definitions Have No Delete Route
    [Documentation]    BR §4.4: a definition is retired by archive; deleting one would orphan the
    ...    counters and issuances that explain every number already handed out.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${SEQ_DEFINITION_API}/${NOT_FOUND_ID}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405

Definitions Have No Bulk Create Or Import Route
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/bulk    json=${{ [] }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    Should Not Be Equal As Integers    ${resp.status_code}    200
