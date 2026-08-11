*** Settings ***
Documentation     Deleting the Product Attribute under test — always the LAST suite, doubling
...               as cleanup. Attribute values referencing the attribute must go first: the FK
...               is ON DELETE NO ACTION, so an attribute with surviving values cannot be
...               removed.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    delete


*** Test Cases ***
Delete With Referencing Attribute Values Fails
    [Documentation]    attribute_id is required-for-create on an attribute value, so removing
    ...    the attribute out from under one would leave it classified as nothing.
    [Tags]    negative
    Ensure Attribute Value Under Test
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=An attribute still referenced by a value must not be deletable

Delete Succeeds
    [Documentation]    With the referencing value gone the attribute deletes normally.
    Delete Inventory Fixture    ${ATTRIBUTE_VALUE_API}    ATTRIBUTE_VALUE_ID
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    attribute under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
