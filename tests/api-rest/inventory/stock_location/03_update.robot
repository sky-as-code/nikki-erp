*** Settings ***
Documentation     Updating Stock Locations. The success cases run first (they consume and
...               rotate the saved etag); negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Location Under Test
Test Tags         inventory    stock_location    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Location
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $STOCK_LOCATION_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${STOCK_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_LOCATION_ETAG}    ${etag}

Update Location Type Succeeds
    [Documentation]    Reclassifying a location is allowed while it is configuration. It is
    ...    changed back so the later suites still see an internal location.
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    json=${{ {'location_type': 'transit', 'etag': $STOCK_LOCATION_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${STOCK_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_LOCATION_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_location.json    200
    Should Be Equal    ${item}[location_type]    transit
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    json=${{ {'location_type': 'internal', 'etag': $item['etag']} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    IF    $etag is not None    Set Global Variable    ${STOCK_LOCATION_ETAG}    ${etag}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Location
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $STOCK_LOCATION_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${STOCK_LOCATION_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $STOCK_LOCATION_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
