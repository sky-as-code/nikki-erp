*** Settings ***
Documentation     Updating Product Attribute Values. The success cases run first (they consume
...               and rotate the saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Attribute Value Under Test
Test Tags         inventory    product_attribute_value    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Value
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $ATTRIBUTE_VALUE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${ATTRIBUTE_VALUE_ETAG}
    IF    $etag is not None    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${etag}

Update Price Extra Succeeds
    [Documentation]    price_extra is ordinary mutable configuration; a negative value must
    ...    round-trip through update the same as create.
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    json=${{ {'price_extra': '-3.25', 'etag': $ATTRIBUTE_VALUE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${ATTRIBUTE_VALUE_ETAG}
    IF    $etag is not None    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute_value.json    200
    Should Be Equal As Numbers    ${item}[price_extra]    -3.25
    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Value
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $ATTRIBUTE_VALUE_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${ATTRIBUTE_VALUE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $ATTRIBUTE_VALUE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
