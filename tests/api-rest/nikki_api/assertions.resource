*** Settings ***
Documentation     Response assertion keywords — ports of the Bruno collection's
...               "scripts/*.js" modules. Status codes and draft-4 schemas (with
...               additionalProperties:false) are asserted together; error contracts pin
...               the exact key/message/type served by the backend. The inconsistent
...               "common.err_*" vs "common:err_*" separators are server contracts —
...               keep them verbatim.
Library           Collections
Resource          common.resource


*** Keywords ***
Response Status Should Be
    [Arguments]    ${resp}    ${status}
    Should Be Equal As Integers    ${resp.status_code}    ${status}
    ...    msg=Unexpected HTTP status. Response body: ${resp.text}

Response Should Match Schema
    [Documentation]    Generic port of testHttpResponse(schema, status). ${schema} is a dict or a path to a .json file.
    [Arguments]    ${resp}    ${schema}    ${status}=200
    Response Status Should Be    ${resp}    ${status}
    Validate Json Schema    ${resp.json()}    ${schema}

Item Should Match Schema
    [Documentation]    Like Response Should Match Schema, but for the single-entity envelope
    ...    {item, meta} the dynamic resource engine returns from GET by id. ${schema}
    ...    describes the entity itself, so only the `item` payload is validated.
    [Arguments]    ${resp}    ${schema}    ${status}=200
    Response Status Should Be    ${resp}    ${status}
    Validate Json Schema    ${resp.json()}[item]    ${schema}
    RETURN    ${resp.json()}[item]

# --- Success contracts ---

Response Should Be Create Success
    [Documentation]    201 with exactly {id, created_at, etag}. Returns (id, etag).
    [Arguments]    ${resp}
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/create_success.json    201
    RETURN    ${resp.json()}[id]    ${resp.json()}[etag]

Response Should Be Update Success
    [Documentation]    200 with {affected_count==count, affected_at, etag?}; the new etag
    ...    must differ from ${previous_etag} when given. Returns the new etag (or None).
    [Arguments]    ${resp}    ${count}=1    ${previous_etag}=${None}
    ${schema}=    Mutate Success Schema    ${count}
    Response Should Match Schema    ${resp}    ${schema}    200
    ${etag}=    Evaluate    $resp.json().get('etag')
    IF    $previous_etag is not None
        Should Not Be Equal    ${etag}    ${previous_etag}
        ...    msg=Response 'etag' must differ from the previous one
    END
    RETURN    ${etag}

Response Should Be Delete Success
    [Arguments]    ${resp}    ${count}=1
    ${schema}=    Delete Success Schema    ${count}
    Response Should Match Schema    ${resp}    ${schema}    200

Response Should Be Exists Success
    [Arguments]    ${resp}    ${existing}    ${not_existing}
    ${schema}=    Exists Success Schema    ${existing}    ${not_existing}
    Response Should Match Schema    ${resp}    ${schema}    200

Response Should Be Search Success
    [Documentation]    200 with the {items, total, page, size} envelope. Items validate
    ...    against ${model_schema} properties (lenient, because of column projections).
    ...    ${item_count}=None only asserts a non-empty result, mirroring testSearch().
    [Arguments]    ${resp}    ${model_schema}    ${size}    ${page}=0    ${item_count}=${None}
    ${schema}=    Search Success Schema    ${model_schema}
    Response Should Match Schema    ${resp}    ${schema}    200
    IF    $item_count is not None
        Length Should Be    ${resp.json()}[items]    ${item_count}
    ELSE
        Should Not Be Empty    ${resp.json()}[items]    msg=Search returned no items
    END
    Should Be Equal As Integers    ${resp.json()}[size]    ${size}
    Should Be Equal As Integers    ${resp.json()}[page]    ${page}

# --- Error contracts (400 unless noted) ---

Response Should Be Missing Fields Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common.err_missing_required_field    field is required
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Invalid Format Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common.err_invalid_data_type    invalid data type, must be a {{typeName}}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Invalid String Length Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common:err_invalid_string_length    string length must be between {{min}} and {{max}}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Invalid Array Length Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common:err_invalid_array_length    array length must be between {{min}} and {{max}}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Invalid Number Range Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common:err_invalid_number_range    value must be between {{min}} and {{max}}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Nonexist Fields Error
    [Arguments]    ${resp}    @{fields}
    ${schema}=    Same Error Schema    ${fields}
    ...    common:err_unknown_schema_field    field is not defined on this schema
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Malformed Payload Error
    [Arguments]    ${resp}
    ${schema}=    One Error Schema No Field    common:err_malformed_request    malformed request
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Etag Unmatched Error
    [Arguments]    ${resp}
    ${schema}=    Same Error Schema    ${{['etag']}}
    ...    common:err_etag_mismatched    This data has been modified by someone else    business
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Not Found Error
    [Arguments]    ${resp}
    ${schema}=    One Error Schema No Field
    ...    common:err_not_found    The desired data could not be found    business
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Field Not Found Error
    [Arguments]    ${resp}    ${field}
    ${schema}=    Same Error Schema    ${{[$field]}}
    ...    common:err_not_found    The desired data could not be found    business
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Value Not Found Error
    [Arguments]    ${resp}    @{values}
    ${vars_schema}=    Evaluate    {'type': 'object', 'required': ['values'], 'properties': {'values': {'type': 'array', 'minItems': len($values), 'maxItems': len($values), 'items': {'type': 'string', 'enum': list($values)}}}}
    ${schema}=    One Error Schema No Field
    ...    common:err_value_not_found    Value(s) could not be found: {.values}    business    ${vars_schema}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Duplicate Values Error
    [Arguments]    ${resp}    @{uniques}
    ${vars_schema}=    Evaluate    {'type': 'object', 'properties': {'uniques': {'type': 'array'}}}
    ${schema}=    One Error Schema No Field
    ...    common.err_unique_constraint_violated    unique constraint violated {{uniques}}    business    ${vars_schema}
    Response Should Match Schema    ${resp}    ${schema}    400
    # `uniques` is a list of unique key GROUPS ([["email"]]), since a composite unique
    # collides as a whole. Callers name the colliding columns, so flatten before comparing.
    ${reported}=    Evaluate    [col for group in $resp.json()[0]['vars']['uniques'] for col in group]
    Lists Should Be Equal    ${reported}    ${uniques}
    ...    msg=Duplicate error reported wrong unique fields

Response Should Be Overlapping Fields Error
    [Arguments]    ${resp}    @{fields}
    ${vars_schema}=    Evaluate    {'type': 'object', 'required': ['fields'], 'properties': {'fields': {'type': 'array', 'minItems': len($fields), 'maxItems': len($fields), 'items': {'type': 'string', 'enum': list($fields)}}}}
    ${schema}=    One Error Schema No Field
    ...    common.err_overlapped_fields    These fields must not have overlapping values: {.fields}    business    ${vars_schema}
    Response Should Match Schema    ${resp}    ${schema}    400

Response Should Be Invalid Access Token Error
    [Arguments]    ${resp}
    ${schema}=    One Error Schema No Field
    ...    common:err_invalid_access_token    Invalid or expired access token.    authorization
    Response Should Match Schema    ${resp}    ${schema}    401
