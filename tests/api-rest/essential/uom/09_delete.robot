*** Settings ***
Documentation     Deleting the Unit of Measure under test — always the LAST suite,
...               doubling as cleanup.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Under Test
Test Tags         essential    uom    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${UOM_API}/${UOM_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so
    ...    the unit under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOM_API}/${UOM_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${UOM_ID}    ${EMPTY}
    Set Global Variable    ${UOM_ETAG}    ${EMPTY}
    Set Global Variable    ${UOM_SYMBOL}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOM_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOM_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
