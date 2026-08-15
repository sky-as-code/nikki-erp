*** Settings ***
Documentation     Archiving the Inventory Location under test, rotating the saved etag. The
...               location is unarchived again so the later suites see it live.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Location Under Test
Test Tags         inventory    inventory_location    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVENTORY_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}

Archived Location Is Still Readable
    [Documentation]    BR §3.2: archive is visibility, not deletion. Historical movements
    ...    still reference this location and must be able to name it.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVENTORY_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
