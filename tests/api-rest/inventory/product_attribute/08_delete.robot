*** Settings ***
Documentation     Deleting the Product Attribute under test — always the LAST suite, doubling
...               as cleanup. An attribute no template ever used takes its values with it in
...               one transaction; one with template history is refused (09_values_and_rules).
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    delete


*** Test Cases ***
Delete Removes The Attribute And Its Unused Values
    [Documentation]    PROD-006 (CR01 §7.5): deleting an attribute no template ever used removes
    ...    its values in the same transaction, so a value is never left classified as nothing.
    Ensure Attribute Value Under Test
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${ATTRIBUTE_VALUE_ID}    ${EMPTY}

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    attribute under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${NOT_FOUND_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/not-existing-1234567890123
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
