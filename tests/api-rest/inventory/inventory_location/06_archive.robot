*** Settings ***
Documentation     Archiving the Inventory Location under test, rotating the saved etag. The
...               location is unarchived again so the later suites see it live.
...               The child created by 02_create.robot is archived first, because a location
...               with a live child refuses to be archived. It is archived rather than deleted
...               so 08_delete.robot still has a child that references the location.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Location Under Test
...               AND    Archive Child Location
Test Tags         inventory    inventory_location    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVENTORY_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}

Archived Location Is Still Readable
    [Documentation]    BR §3.2: archive is visibility, not deletion. Historical movements
    ...    still reference this location and must be able to name it.
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${INVENTORY_LOCATION_ETAG}
    IF    $etag is not None    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $INVENTORY_LOCATION_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived


*** Keywords ***
Archive Child Location
    ${child}=    Get Variable Value    ${CHILD_INVENTORY_LOCATION_ID}    ${EMPTY}
    IF    not $child    RETURN
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${child}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${child}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'is_archived': True, 'etag': $etag} }}
    Response Status Should Be    ${resp}    200
