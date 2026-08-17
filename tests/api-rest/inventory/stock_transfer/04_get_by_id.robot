*** Settings ***
Documentation     Reading one Stock Transfer.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Transfer Under Test
Test Tags         inventory    stock_transfer    read


*** Test Cases ***
Get By Id Returns The Transfer
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${STOCK_TRANSFER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_transfer.json    200
    Should Be Equal    ${item}[id]    ${STOCK_TRANSFER_ID}
    Should Be Equal    ${item}[source_location_id]    ${INVENTORY_LOCATION_ID}
    Should Be Equal    ${item}[destination_location_id]    ${STOCK_DEST_LOCATION_ID}

Get By Id Of An Unknown Transfer Returns Not Found
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${NOT_FOUND_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
