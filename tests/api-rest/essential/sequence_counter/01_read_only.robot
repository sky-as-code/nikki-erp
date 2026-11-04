*** Settings ***
Documentation     Counters are listed and read, never written by a client (BR §8.3): no create,
...               update, archive or delete route. Large integers travel as decimal strings
...               (BR §6.1, AC37) and a page holds at most 200 rows.
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Issued Sequence Definition
Test Tags         essential    sequence_counter


*** Test Cases ***
Get Sequence Counter Model Schema
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

The Operation Type Has Its Own Counter On The Org Definition
    [Documentation]    SEQ-FB-002: the org's one inventory.transfer definition keeps a counter per
    ...    operation type (the scope object), period NONE (never resets), standing at the one
    ...    number the fresh type issued.
    ${resp}=    Search Sequence Records    ${SEQ_COUNTER_API}    scope_object_id    ${SEQ_ISSUED_TYPE_ID}
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_counter.json    50    item_count=1
    ${counter}=    Set Variable    ${resp.json()}[items][0]
    Should Be Equal    ${counter}[definition_id]    ${SEQ_ISSUED_DEFINITION_ID}
    Should Be Equal    ${counter}[last_number]    1
    Should Be Equal    ${counter}[period_key]    NONE
    Should Be Equal    ${counter}[org_id]    ${INV_ORG_ID}
    Response Should Carry No Secret    ${resp}
    Set Suite Variable    ${COUNTER_ID}    ${counter}[id]

Get By Id Returns The Counter With A String Last Number
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}/${COUNTER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_counter.json    200
    Should Be Equal    ${item}[definition_id]    ${SEQ_ISSUED_DEFINITION_ID}
    Should Be String    ${item}[last_number]
    Should Be Equal    ${item}[scope_object_id]    ${SEQ_ISSUED_TYPE_ID}
    Response Should Carry No Secret    ${resp}

The Next Transfer Advances The Counter By One Step
    ${transfer_id}    ${etag}=    Create Stock Transfer    ${SEQ_ISSUED_TYPE_ID}
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}/${COUNTER_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[item][last_number]    2
    [Teardown]    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

A Page Holds At Most 200 Counters
    [Documentation]    BR §8.3: a larger size is capped, not refused.
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}    params=${{ {'size': 500} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    200
    ${count}=    Get Length    ${resp.json()}[items]
    Should Be True    ${count} <= 200

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}    params=${{ {'page': 0, 'size': 1} }}
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_counter.json    1    item_count=1

Filtering On The Secret Field Fails
    [Tags]    negative
    ${resp}=    Search Sequence Records    ${SEQ_COUNTER_API}    algorithm_params    x
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_READ_ONLY_FIELD}    graph

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${SEQ_COUNTER_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Counters Have No Write Routes
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_COUNTER_API}    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    PATCH On Session    api    ${SEQ_COUNTER_API}/${COUNTER_ID}
    ...    json=${{ {'last_number': '0'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    DELETE On Session    api    ${SEQ_COUNTER_API}/${COUNTER_ID}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    POST On Session    api    ${SEQ_COUNTER_API}/${COUNTER_ID}/archived
    ...    json=${{ {'is_archived': True} }}    expected_status=any
    Should Be True    ${resp.status_code} in (404, 405)
