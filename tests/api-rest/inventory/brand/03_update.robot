*** Settings ***
Documentation     Updating Brands. The success cases run first (they consume and rotate the
...               saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Brand Under Test
Test Tags         inventory    brand    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Brand
    ${resp}=    PATCH On Session    api    ${BRAND_API}/${BRAND_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $BRAND_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${BRAND_ETAG}
    IF    $etag is not None    Set Global Variable    ${BRAND_ETAG}    ${etag}

Update Website Succeeds
    [Documentation]    website is an ordinary mutable optional field, same as any other.
    ${resp}=    PATCH On Session    api    ${BRAND_API}/${BRAND_ID}
    ...    json=${{ {'website': 'https://updated.example.com', 'etag': $BRAND_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${BRAND_ETAG}
    IF    $etag is not None    Set Global Variable    ${BRAND_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${BRAND_API}/${BRAND_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/brand.json    200
    Should Be Equal    ${item}[website]    https://updated.example.com
    Set Global Variable    ${BRAND_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${BRAND_API}/${BRAND_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Brand
    ${resp}=    PATCH On Session    api    ${BRAND_API}/${BRAND_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${BRAND_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $BRAND_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${BRAND_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $BRAND_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
