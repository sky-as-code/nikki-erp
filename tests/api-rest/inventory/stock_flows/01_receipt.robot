*** Settings ***
Documentation     The receipt path: create -> confirm -> validate -> on-hand increased.
...
...               This is the flow every other one depends on, because Phase 1 made the quant
...               read-only to clients: validating an incoming transfer is the only way stock can
...               enter the system at all (AC-STOCK-002).
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Product Variant Under Test
Test Tags         inventory    stock_flows    receipt


*** Test Cases ***
Receipt Raises On Hand By The Processed Quantity
    [Documentation]    AC-STOCK-006. The destination balance goes up by exactly what was
    ...    processed — not by the demand, and not by a rounded figure.
    ${before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}

    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    25
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    ${expected}=    Evaluate    ${before} + 25
    Should Be Equal As Numbers    ${after}    ${expected}
    ...    msg=A validated receipt must raise on-hand by exactly the quantity it processed

    Set Suite Variable    ${RECEIPT_TRANSFER_ID}    ${transfer_id}
    Set Suite Variable    ${RECEIPT_MOVE_ID}    ${move_id}

Validated Receipt Is Done And Stamped
    [Documentation]    BR §4.2.3.10 postconditions: the transfer is Done and completed_at is
    ...    recorded. The stamp is what makes it historical rather than merely finished.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${RECEIPT_TRANSFER_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    done
    Should Not Be Equal    ${item.get('completed_at')}    ${None}
    ...    msg=completed_at must be stamped by validate

Validated Receipt Closes Its Moves
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}/${RECEIPT_MOVE_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[data][status]    done

Validating A Done Transfer Fails
    [Documentation]    Validate is not repeatable without an idempotency key: a second run would
    ...    move the stock again.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${RECEIPT_TRANSFER_ID}/validate
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A completed transfer must not be validated a second time

Cancelling A Done Transfer Is Refused
    [Documentation]    AC-STOCK-009 and STOCK-INV-005. The goods physically moved; no sequence of
    ...    edits can make that not have happened, so the remedy is a reverse transfer.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${RECEIPT_TRANSFER_ID}/cancel
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A completed transfer cannot be cancelled

Validating Does Not Reserve Anything At The Destination
    [Documentation]    Goods that have arrived are available, not spoken for. A receipt that left
    ...    them reserved would hide them from every later demand.
    ${reserved}=    Read Stock Reserved    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}
    Should Be Equal As Numbers    ${reserved}    0
    ...    msg=Received stock must arrive unreserved
