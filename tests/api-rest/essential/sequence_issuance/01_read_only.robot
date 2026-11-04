*** Settings ***
Documentation     Issuances are listed and read, never written by a client. Each carries the
...               idempotency fields that make a retried request replay instead of re-issue
...               (BR §6.3): idempotency id, operation key, fingerprint, and its expiry.
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Issued Sequence Definition
Test Tags         essential    sequence_issuance


*** Test Cases ***
Get Sequence Issuance Model Schema
    ${resp}=    GET On Session    api    ${SEQ_ISSUANCE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

The Transfer Left Exactly One Issuance On Its Type's Counter
    ${counter}=    Type Sequence Counter    ${SEQ_ISSUED_TYPE_ID}
    Set Suite Variable    ${COUNTER_ID}    ${counter}[id]
    ${resp}=    Search Sequence Records    ${SEQ_ISSUANCE_API}    counter_id    ${COUNTER_ID}
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_issuance.json    50    item_count=1
    Should Be Equal    ${resp.json()}[items][0][definition_id]    ${SEQ_ISSUED_DEFINITION_ID}
    Set Suite Variable    ${ISSUANCE_ID}    ${resp.json()}[items][0][id]

The Issuance Records Who Asked, For What And Until When
    [Documentation]    The operation key names the field it numbered and nothing minted per attempt
    ...    (SEQ-FB-001): with the request's idempotency id it is what a retry replays by. The caller
    ...    is Inventory, the value is the transfer number's counter segment, and the replay window is
    ...    the definition's retention (90 days by default).
    ${resp}=    GET On Session    api    ${SEQ_ISSUANCE_API}/${ISSUANCE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_issuance.json    200
    Should Be Equal    ${item}[operation_key]    transfer_number
    Should Not Be Empty    ${item}[idempotency_id]
    Should Be Equal    ${item}[caller_module]    inventory
    Should Be Equal    ${item}[allocated_value]    1
    Should Be Equal    ${item}[org_id]    ${INV_ORG_ID}
    Should End With    ${SEQ_ISSUED_TRANSFER_NUMBER}    /${item}[formatted_value]
    Should Not Be Empty    ${item}[request_fingerprint]
    ${days}=    Evaluate
    ...    (__import__('datetime').datetime.fromisoformat($item['expires_at'].replace('Z', '+00:00')) - __import__('datetime').datetime.fromisoformat($item['issued_at'].replace('Z', '+00:00'))).days
    Should Be Equal As Integers    ${days}    90
    Response Should Carry No Secret    ${resp}

The Issuance Points At The Definition's Counter
    ${resp}=    GET On Session    api    ${SEQ_ISSUANCE_API}/${ISSUANCE_ID}
    Should Be Equal    ${resp.json()}[item][counter_id]    ${COUNTER_ID}

Every Transfer Gets Its Own Issuance And Number
    [Documentation]    AC13-ish: two requests carry two idempotency ids, so they are independent
    ...    actions under the same operation key and two transfers never share a number.
    ${transfer_id}    ${etag}=    Create Stock Transfer    ${SEQ_ISSUED_TYPE_ID}
    ${resp}=    Search Sequence Records    ${SEQ_ISSUANCE_API}    counter_id    ${COUNTER_ID}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    2
    ${values}=    Evaluate    sorted(int(item['allocated_value']) for item in $resp.json()['items'])
    Should Be Equal    ${values}    ${{ [1, 2] }}
    [Teardown]    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

A Retry With The Same Idempotency Id Issues No Second Number
    [Documentation]    SEQ-FB-001: the same X-Idempotency-ID and the same operation key replay the
    ...    issuance, so a retried create gets the number already handed out, collides on the per-org
    ...    unique transfer number and is refused instead of becoming a second document.
    ${before}=    Search Sequence Records    ${SEQ_ISSUANCE_API}    counter_id    ${COUNTER_ID}
    ${count}=    Get Length    ${before.json()}[items]
    ${idempotency_id}=    New Idempotency Id
    ${body}=    Set Variable    ${{ {'operation_type_id': $SEQ_ISSUED_TYPE_ID, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${headers}=    Set Variable    ${{ {'X-Idempotency-ID': $idempotency_id} }}
    ${first}=    POST On Session    api    ${STOCK_TRANSFER_API}    json=${body}    headers=${headers}
    ${transfer_id}    ${etag}=    Response Should Be Create Success    ${first}
    ${second}=    POST On Session    api    ${STOCK_TRANSFER_API}    json=${body}    headers=${headers}
    ...    expected_status=any
    IF    $second.status_code == 201
        Should Be Equal    ${second.json()}[id]    ${transfer_id}
        ...    msg=A replayed create must not become a second transfer
    END
    ${after}=    Search Sequence Records    ${SEQ_ISSUANCE_API}    counter_id    ${COUNTER_ID}
    Length Should Be    ${after.json()}[items]    ${{ $count + 1 }}
    [Teardown]    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Search By Caller Module Succeeds
    ${resp}=    Search Sequence Records    ${SEQ_ISSUANCE_API}    caller_module    inventory    size=5
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_issuance.json    5
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[caller_module]    inventory
    END

A Page Holds At Most 200 Issuances
    ${resp}=    GET On Session    api    ${SEQ_ISSUANCE_API}    params=${{ {'size': 1000} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    200

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${SEQ_ISSUANCE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Issuances Have No Write Routes
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_ISSUANCE_API}    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    PATCH On Session    api    ${SEQ_ISSUANCE_API}/${ISSUANCE_ID}
    ...    json=${{ {'formatted_value': 'X'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    DELETE On Session    api    ${SEQ_ISSUANCE_API}/${ISSUANCE_ID}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
