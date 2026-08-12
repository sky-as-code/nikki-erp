*** Settings ***
Documentation     The transfer state machine over HTTP (BR §6.1, §4.2.3.6, §4.2.3.12).
...
...               These are the rules that hold regardless of stock: what may be confirmed, what
...               may be cancelled, and what a completed transfer refuses. The paths that move
...               real quantities live in inventory/stock_flows.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
Test Tags         inventory    stock_transfer    lifecycle


*** Test Cases ***
Confirming A Transfer With No Moves Fails
    [Documentation]    BR §4.2.3.6. Such a transfer can never become ready and has nothing to
    ...    validate, so confirming it would produce a document stuck forever.
    [Tags]    negative
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A transfer with no moves must not be confirmable
    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Confirming A Draft Transfer Succeeds
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${check}=    GET On Session    api    ${STOCK_TRANSFER_API}/${id}
    Should Not Be Equal    ${check.json()}[data][status]    draft
    ...    msg=Confirm must take the transfer out of draft
    Set Test Variable    ${CONFIRMED_TRANSFER_ID}    ${id}
    Set Test Variable    ${CONFIRMED_MOVE_ID}    ${move_id}

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Confirming Does Not Change On Hand
    [Documentation]    AC-STOCK-004. Confirming commits to the demand, not to any stock having
    ...    moved; only validate changes an on-hand quantity.
    ${before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${before}
    ...    msg=Confirming a transfer must not change any on-hand quantity

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Confirming A Confirmed Transfer Fails
    [Documentation]    Confirm applies to a draft. Re-confirming would re-run the reservation
    ...    step against a transfer already past it.
    [Tags]    negative
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Only a draft transfer can be confirmed

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Cancelling A Draft Transfer Succeeds
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${check}=    GET On Session    api    ${STOCK_TRANSFER_API}/${id}
    Should Be Equal    ${check.json()}[data][status]    cancelled

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Cancelling Closes The Transfer's Moves
    [Documentation]    A cancelled transfer must not leave open moves behind: they would still
    ...    look reservable, against a document nobody is going to process.
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel    json=${{ {} }}    expected_status=any

    ${check}=    GET On Session    api    ${STOCK_MOVE_API}/${move_id}
    Should Be Equal    ${check.json()}[data][status]    cancelled
    ...    msg=Cancelling a transfer must cancel its open moves

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Cancelling A Cancelled Transfer Fails
    [Tags]    negative
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    5
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel    json=${{ {} }}    expected_status=any

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Cancelled is terminal

    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}    expected_status=any
    ...    AND    DELETE On Session    api    ${STOCK_TRANSFER_API}/${id}    expected_status=any

Operating On An Unknown Transfer Fails
    [Tags]    negative
    FOR    ${action}    IN    confirm    reserve    unreserve    validate    cancel
        ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${NOT_FOUND_ID}/${action}
        ...    json=${{ {} }}    expected_status=any
        Should Not Be Equal As Integers    ${resp.status_code}    200
        ...    msg=${action} on a transfer that does not exist must not report success
    END
