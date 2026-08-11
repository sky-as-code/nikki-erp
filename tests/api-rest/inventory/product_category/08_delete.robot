*** Settings ***
Documentation     Deleting the Product Category under test — always the LAST suite, doubling
...               as cleanup. The child category referencing it as parent must go first: the
...               self-FK is ON DELETE NO ACTION, so a category with a surviving child cannot
...               be removed.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Category Under Test
Test Tags         inventory    product_category    delete


*** Test Cases ***
Delete With Referencing Child Fails
    [Documentation]    parent_category_id is a self-FK declared ON DELETE NO ACTION, so
    ...    removing a category out from under its child would leave the child pointing at
    ...    nothing.
    [Tags]    negative
    Ensure Child Product Category
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A category still referenced by a child must not be deletable

Delete Succeeds
    [Documentation]    With the child gone the parent deletes normally. The child goes
    ...    before the category under test.
    Delete Inventory Fixture    ${PRODUCT_CATEGORY_API}    CHILD_CATEGORY_ID
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    category under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_CATEGORY_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
