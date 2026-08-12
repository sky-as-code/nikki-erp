*** Settings ***
Documentation     Deleting Product Variants — always the LAST suite, doubling as cleanup.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_variant    delete


*** Test Cases ***
Delete Succeeds
    [Documentation]    An unreferenced variant deletes normally. BR-PROD-VAR-008 blocks the
    ...    hard delete only once a transaction references it, which is a state this suite
    ...    cannot construct without a vending-machine order.
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    variant under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_VARIANT_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${EMPTY}

Deleting A Variant Frees Its Combination
    [Documentation]    BR-PROD-VAR-002 constrains live rows only: once a variant is gone its
    ...    combination is available again, or a deleted mistake would poison that
    ...    combination for the template forever.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Recycle Template
    ${key}=    Unique Code    recycle
    ${first_id}    ${first_etag}=    Create Product Variant    ${template_id}    ${key}
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${first_id}
    ${second_id}    ${second_etag}=    Create Product Variant    ${template_id}    ${key}
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${second_id}    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}    expected_status=any

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
