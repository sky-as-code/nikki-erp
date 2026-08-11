*** Settings ***
Documentation     Deleting the Product Template under test — always the LAST suite, doubling
...               as cleanup. The delete guard of BR-PROD-TPL-005 / AC-PROD-021 is the point:
...               a template whose variants may already be referenced by a transaction must
...               be archived rather than deleted, and the error must say so.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_template    delete


*** Test Cases ***
Delete With Variants Fails
    [Documentation]    BR-PROD-TPL-005 / AC-PROD-021: hard-deleting a template with variants
    ...    would break the identity of every transaction line referencing one. The engine
    ...    turns this into a field-level business error suggesting Archive, rather than
    ...    letting the FK surface as a 500.
    [Tags]    negative
    Ensure Product Variant Under Test
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    expected_status=any
    Response Should Be Template Has Variants Error    ${resp}

Delete Succeeds
    [Documentation]    With its variants gone the template deletes normally — the guard is
    ...    about surviving references, not about templates being undeletable.
    Delete Inventory Fixture    ${PRODUCT_VARIANT_API}    PRODUCT_VARIANT_ID
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    template under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_TEMPLATE_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
