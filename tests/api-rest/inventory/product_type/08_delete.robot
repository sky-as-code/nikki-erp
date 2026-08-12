*** Settings ***
Documentation     Deleting the Product Type under test — always the LAST suite, doubling as
...               cleanup. Templates referencing the type must go first: the FK is
...               ON DELETE NO ACTION, so a type with surviving templates cannot be removed.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Type Under Test
Test Tags         inventory    product_type    delete


*** Test Cases ***
Delete With Referencing Templates Fails
    [Documentation]    product_type_id is required-for-create on a template, so removing the
    ...    type out from under one would leave it classified as nothing.
    [Tags]    negative
    Ensure Product Template Under Test
    ${resp}=    DELETE On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A product type still referenced by a template must not be deletable

Delete Succeeds
    [Documentation]    With the referencing records gone the type deletes normally. The
    ...    variant goes before its template, and the template before the type.
    Delete Inventory Fixture    ${PRODUCT_VARIANT_API}    PRODUCT_VARIANT_ID
    Delete Inventory Fixture    ${PRODUCT_TEMPLATE_API}    PRODUCT_TEMPLATE_ID
    ${resp}=    DELETE On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    type under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_TYPE_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TYPE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TYPE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
