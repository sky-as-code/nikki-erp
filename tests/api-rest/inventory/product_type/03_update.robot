*** Settings ***
Documentation     Updating Product Types. The success case runs first (it consumes and
...               rotates the saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Type Under Test
Test Tags         inventory    product_type    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Type
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $PRODUCT_TYPE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}

Update Capability Flags Succeeds
    [Documentation]    BR §6.3.2: the flags are ordinary mutable configuration. Turning
    ...    stock support off is how an existing type is reclassified as a service.
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ...    json=${{ {'supports_manufacturing': True, 'etag': $PRODUCT_TYPE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_type.json    200
    Should Be Equal    ${item}[supports_manufacturing]    ${True}
    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Type
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/${PRODUCT_TYPE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PRODUCT_TYPE_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TYPE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PRODUCT_TYPE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
