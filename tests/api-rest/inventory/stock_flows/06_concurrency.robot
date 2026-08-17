*** Settings ***
Documentation     Concurrent reservation against one balance (AC-STOCK-007, BR §8.6).
...
...               This is the test the row lock exists for. Two requests reserve against the same
...               quant at the same time; both read the same available quantity if nothing
...               serialises them, and both then reserve it — leaving more stock promised than
...               exists, with no error anywhere to show for it.
...
...               The assertion is deliberately about the total rather than about which request
...               won. Either outcome is legitimate; what is not legitimate is both of them
...               succeeding in full.
Library           Collections
Library           Process
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
Test Tags         inventory    stock_flows    concurrency


*** Test Cases ***
Two Reservations Against One Balance Do Not Over Reserve
    [Documentation]    Both transfers demand the whole available quantity, so between them they
    ...    ask for twice what exists. However the two interleave, the reserved total must never
    ...    exceed what is on hand: reserved beyond on-hand is stock promised to someone that is
    ...    not there to give.
    Seed Exact Stock For Contention

    ${first_id}    ${first_etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${first_move}=    Add Stock Move    ${first_id}    ${PRODUCT_VARIANT_ID}    ${CONTENDED_QUANTITY}
    ${second_id}    ${second_etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${second_move}=    Add Stock Move    ${second_id}    ${PRODUCT_VARIANT_ID}    ${CONTENDED_QUANTITY}

    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/confirm    json=${{ {} }}    expected_status=any

    #    Fired back to back rather than truly in parallel: RequestsLibrary is synchronous, so this
    #    is the closest a Robot suite gets without a second process. It still catches the failure
    #    it is aimed at, because the bug it guards is a read-then-write window that survives any
    #    ordering — if availability were read before the lock, the second request would allocate
    #    against the figure the first had already spent.
    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/reserve    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/reserve    json=${{ {} }}    expected_status=any

    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be True    ${reserved} <= ${on_hand}
    ...    msg=Reserved (${reserved}) must never exceed on-hand (${on_hand}): two requests both reserved the same stock

    [Teardown]    Run Keywords
    ...    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${first_move}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${second_move}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${first_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${second_id}    expected_status=any

Exactly One Of Two Contenders Is Fully Assigned
    [Documentation]    The stronger form of the same rule. With exactly one demand's worth of
    ...    stock available, one transfer should get all of it and the other should not: two
    ...    fully-assigned moves would mean the same units were promised twice.
    Seed Exact Stock For Contention

    ${first_id}    ${first_etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${first_move}=    Add Stock Move    ${first_id}    ${PRODUCT_VARIANT_ID}    ${CONTENDED_QUANTITY}
    ${second_id}    ${second_etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${second_move}=    Add Stock Move    ${second_id}    ${PRODUCT_VARIANT_ID}    ${CONTENDED_QUANTITY}

    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/reserve    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/reserve    json=${{ {} }}    expected_status=any

    ${first}=    GET On Session    api    ${STOCK_MOVE_API}/${first_move}
    ${second}=    GET On Session    api    ${STOCK_MOVE_API}/${second_move}
    ${first_status}=    Set Variable    ${first.json()}[data][status]
    ${second_status}=    Set Variable    ${second.json()}[data][status]
    ${assigned}=    Evaluate    [$first_status, $second_status].count('assigned')
    Should Be True    ${assigned} <= 1
    ...    msg=Both moves reported as assigned (${first_status}, ${second_status}): the same stock was promised twice

    [Teardown]    Run Keywords
    ...    POST On Session    api    ${STOCK_TRANSFER_API}/${first_id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    POST On Session    api    ${STOCK_TRANSFER_API}/${second_id}/cancel    json=${{ {} }}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${first_move}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_MOVE_API}/${second_move}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${first_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${second_id}    expected_status=any


*** Keywords ***
Seed Exact Stock For Contention
    [Documentation]    Records how much is on hand and makes the contended demand equal to it, so
    ...    that two transfers asking for it are guaranteed to overlap. Receives stock first when
    ...    the location is empty, since a contention test over zero stock proves nothing.
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${free}=    Evaluate    ${on_hand} - ${reserved}
    IF    ${free} < 10
        ${transfer_id}    ${move_id}=    Receive Stock Into Location
        ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    60
        POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
        ...    json=${{ {} }}    expected_status=any
        ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
        ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
        ${free}=    Evaluate    ${on_hand} - ${reserved}
    END
    ${contended}=    Evaluate    int(${free})
    Set Suite Variable    ${CONTENDED_QUANTITY}    ${contended}
