*** Settings ***
Documentation     Deleting the Stock Operation Type under test — always the LAST suite,
...               doubling as cleanup. No transfer exists yet in this phase, so the type has
...               no dependant and deletes cleanly. Once transfers land, deleting a type a
...               transfer references must be refused in favour of archive (BR §4.2.1.5).
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    type under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ...    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ID}    ${EMPTY}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/not-existing-1234567890123
    ...    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
