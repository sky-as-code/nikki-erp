*** Settings ***
Documentation     Updating Stock Operation Types. The success cases run first (they consume
...               and rotate the saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Operation
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $STOCK_OPERATION_TYPE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    ...    previous_etag=${STOCK_OPERATION_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}

Update Reservation Method Succeeds
    [Documentation]    BR §4.2.1.4: policy is reconfigurable while the type is not archived.
    ...    A transfer already created keeps the policy it snapshotted.
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    json=${{ {'reservation_method': 'manual', 'backorder_policy': 'always', 'etag': $STOCK_OPERATION_TYPE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    ...    previous_etag=${STOCK_OPERATION_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ${item}=    Item Should Match Schema    ${resp}
    ...    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[reservation_method]    manual
    Should Be Equal    ${item}[backorder_policy]    always
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${item}[etag]

Update Operation Code Is Rejected Or Ignored
    [Documentation]    operation_code is declared no_update. The engine drops a no_update
    ...    field rather than failing the request, so this pins the stored value is unchanged
    ...    either way — what must never happen is the direction silently flipping.
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    json=${{ {'operation_code': 'outgoing', 'etag': $STOCK_OPERATION_TYPE_ETAG} }}
    ...    expected_status=any
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ${item}=    Item Should Match Schema    ${resp}
    ...    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[operation_code]    incoming
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Operation
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_OPERATION_TYPE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $STOCK_OPERATION_TYPE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
