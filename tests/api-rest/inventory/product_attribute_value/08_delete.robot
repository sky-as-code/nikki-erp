*** Settings ***
Documentation     Deleting the Product Attribute Value under test — always the LAST suite,
...               doubling as cleanup.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Attribute Value Under Test
Test Tags         inventory    product_attribute_value    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    value under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${ATTRIBUTE_VALUE_ID}    ${EMPTY}
    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
