*** Settings ***
Documentation     Updating Product Attributes. The success cases run first (they consume and
...               rotate the saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Attribute
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $PRODUCT_ATTRIBUTE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_ATTRIBUTE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${etag}

Update Variant Creation Mode Succeeds
    [Documentation]    BR §14.3 step 2: variant_creation_mode is ordinary mutable
    ...    configuration. Switching an attribute to `dynamic` is how it is reclassified so
    ...    its combinations materialize lazily instead of eagerly.
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    json=${{ {'variant_creation_mode': 'dynamic', 'etag': $PRODUCT_ATTRIBUTE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_ATTRIBUTE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute.json    200
    Should Be Equal    ${item}[variant_creation_mode]    dynamic
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${item}[etag]

Update With Invalid Variant Creation Mode Fails
    [Documentation]    The enum stays closed on update, not just on create.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    json=${{ {'variant_creation_mode': 'sometimes', 'etag': $PRODUCT_ATTRIBUTE_ETAG} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=variant_creation_mode must reject a value outside instant/dynamic/never

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Attribute
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PRODUCT_ATTRIBUTE_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_ATTRIBUTE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PRODUCT_ATTRIBUTE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
