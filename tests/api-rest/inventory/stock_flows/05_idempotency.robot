*** Settings ***
Documentation     Replayed validation (BR §8.7).
...
...               This guards the one failure in the module that cannot be repaired afterwards. A
...               client whose request times out retries; if the first attempt had in fact
...               succeeded, the retry must not ship the goods a second time. An edit can fix a
...               wrong number in a record — nothing can un-ship a second delivery.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Seed Stock For Idempotency
Test Tags         inventory    stock_flows    idempotency


*** Test Cases ***
Replayed Validate Does Not Move Stock Twice
    ${key}=    Unique Code    idem

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    15
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve    json=${{ {} }}    expected_status=any

    ${source_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${dest_before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}

    ${first}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {'idempotency_key': $key} }}    expected_status=any
    Response Status Should Be    ${first}    200

    ${source_once}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${dest_once}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}

    #    The retry a timed-out client sends: same transfer, same key.
    ${second}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {'idempotency_key': $key} }}    expected_status=any
    #    A replayed validate must report the prior success, not an error: the client's first
    #    attempt did in fact work, it just never saw the response.
    Response Status Should Be    ${second}    200

    ${source_twice}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${dest_twice}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_DEST_LOCATION_ID}
    Should Be Equal As Numbers    ${source_twice}    ${source_once}
    ...    msg=The retry must not take from the source a second time
    Should Be Equal As Numbers    ${dest_twice}    ${dest_once}
    ...    msg=The retry must not deliver to the destination a second time

    ${expected_source}=    Evaluate    ${source_before} - 15
    Should Be Equal As Numbers    ${source_once}    ${expected_source}
    ...    msg=The first validate must have moved the stock exactly once

    Set Suite Variable    ${IDEMPOTENT_TRANSFER_ID}    ${id}
    Set Suite Variable    ${IDEMPOTENT_KEY}    ${key}

Validate Records The Key It Completed Under
    [Documentation]    The key is stored in the same transaction as the movements it guards.
    ...    Written afterwards there would be a window in which the stock had moved but a retry
    ...    could not tell.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${IDEMPOTENT_TRANSFER_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[data][idempotency_key]    ${IDEMPOTENT_KEY}

Validate With A Different Key Is Not A Replay
    [Documentation]    A different key is a different operation. It must fall through to the
    ...    already-closed check rather than short-circuiting into a false success.
    [Tags]    negative
    ${other}=    Unique Code    idemx
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${IDEMPOTENT_TRANSFER_ID}/validate
    ...    json=${{ {'idempotency_key': $other} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A done transfer must refuse a validate carrying an unfamiliar key

Validate Without A Key Gets No Replay Protection
    [Documentation]    Idempotency is opt-in. A caller sending no key must not accidentally match
    ...    a transfer that happens to carry one — it gets the ordinary already-closed refusal.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${IDEMPOTENT_TRANSFER_ID}/validate
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200


*** Keywords ***
Seed Stock For Idempotency
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    IF    ${on_hand} >= 50    RETURN
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}    100
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
