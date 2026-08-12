*** Settings ***
Documentation     Deleting Stock Transfers, and the cleanup of the fixture this suite created.
...               Always last: the earlier files depend on the transfer under test.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Transfer Under Test
Test Tags         inventory    stock_transfer    delete


*** Test Cases ***
Delete Removes The Transfer
    [Documentation]    The move goes first: it points at the transfer, and every FK in this
    ...    module is ON DELETE NO ACTION, so the other order leaves the row behind.
    ${move}=    Get Variable Value    ${STOCK_MOVE_ID}    ${EMPTY}
    IF    $move
        DELETE On Session    api    ${STOCK_MOVE_API}/${move}    expected_status=any
        Set Global Variable    ${STOCK_MOVE_ID}    ${EMPTY}
    END

    ${resp}=    DELETE On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ...    expected_status=any
    Should Be True    ${resp.status_code} in (200, 204)
    ...    msg=Deleting a draft transfer with no moves left must succeed
    Set Global Variable    ${STOCK_TRANSFER_ID}    ${EMPTY}

Delete Of An Unknown Transfer Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_TRANSFER_API}/${NOT_FOUND_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
