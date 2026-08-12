*** Settings ***
Documentation     Updating Stock Transfers. What may be edited is narrow by design: a transfer's
...               identity, direction and policies are fixed at create time, so an update reaches
...               the scheduling and annotation fields and little else.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Transfer Under Test
Test Tags         inventory    stock_transfer    update


*** Test Cases ***
Update Note Succeeds
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ...    json=${{ {'note': 'Robot updated note', 'etag': $STOCK_TRANSFER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${STOCK_TRANSFER_ETAG}    ${etag}

Update With A Stale Etag Fails
    [Documentation]    The etag is the concurrency guard: a client editing a copy it fetched
    ...    before someone else's write must be told, not silently allowed to overwrite them.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ...    json=${{ {'note': 'Robot stale write', 'etag': '___________________'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A stale etag must be refused

Update Cannot Change The Status
    [Documentation]    status is no_update: it is derived from the transfer's moves by confirm,
    ...    reserve, validate and cancel. A hand-set status would claim an outcome no movement
    ...    produced, which is the same hole the quant's read-only rule closes.
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ...    json=${{ {'status': 'done', 'etag': $STOCK_TRANSFER_ETAG} }}
    ...    expected_status=any
    ${check}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    Should Not Be Equal    ${check.json()}[data][status]    done
    ...    msg=A transfer must not be driven to done by a plain update
    ${etag}=    Set Variable    ${check.json()}[data][etag]
    Set Global Variable    ${STOCK_TRANSFER_ETAG}    ${etag}

Update Cannot Change The Transfer Number
    [Documentation]    The number is the reference quoted on paperwork and in other documents,
    ...    so renumbering a transfer would orphan every mention of it.
    ${before}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ${original}=    Set Variable    ${before.json()}[data][transfer_number]
    PATCH On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ...    json=${{ {'transfer_number': 'ROBOT-RENUMBERED', 'etag': $STOCK_TRANSFER_ETAG} }}
    ...    expected_status=any
    ${after}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    Should Be Equal    ${after.json()}[data][transfer_number]    ${original}
    ...    msg=transfer_number is immutable
    ${etag}=    Set Variable    ${after.json()}[data][etag]
    Set Global Variable    ${STOCK_TRANSFER_ETAG}    ${etag}

Update Of An Unknown Transfer Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${NOT_FOUND_ID}
    ...    json=${{ {'note': 'Robot ghost', 'etag': '___________________'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
