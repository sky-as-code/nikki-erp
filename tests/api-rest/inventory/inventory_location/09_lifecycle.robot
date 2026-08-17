*** Settings ***
Documentation     Suspend, resume and the archive guard on a location.
...
...               Suspend and archive read the same facts and reach opposite conclusions, on
...               purpose. Locking a damaged rack that still holds goods is exactly what
...               suspension is for, so stock does not block it; archiving a location that
...               still holds stock would strand the goods, so stock does block that. Anyone
...               tidying the two guards into one will break this file, which is the point.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Lifecycle Location
Test Tags         inventory    inventory_location    lifecycle


*** Test Cases ***
Suspend Takes A Location Out Of Use
    [Documentation]    TS-STATUS-04.
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/suspend
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Lifecycle Location
    Should Be Equal    ${item}[status]    suspended
    Should Not Be True    ${item}[is_archived]
    ...    msg=Suspending must not archive: the location is locked, not withdrawn

Suspending A Suspended Location Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/suspend
    ...    json=${{ {} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Resume Returns A Location To Use
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/resume
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Lifecycle Location
    Should Be Equal    ${item}[status]    active

Unarchive Leaves A Location Suspended
    [Documentation]    TS-STATUS-12. Like a warehouse, a location comes back suspended rather
    ...    than active: the tree it sat in may have changed while it was archived.
    ${item}=    Get Lifecycle Location
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Lifecycle Location
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/archived
    ...    json=${{ {'is_archived': False, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Lifecycle Location
    Should Not Be True    ${item}[is_archived]
    Should Be Equal    ${item}[status]    suspended
    ...    msg=Unarchiving returns a location to suspended, never straight to active

    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/resume
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200

Archiving A Location With A Live Child Is Refused
    [Documentation]    Archiving a parent would leave its children hanging off something
    ...    withdrawn from the working set, so the children go first.
    [Tags]    negative
    ${child}=    Create Child Of Lifecycle Location
    ${item}=    Get Lifecycle Location
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A location with a live child must not archive
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${child}    expected_status=any

A System Generated Location Cannot Be Archived While Its Warehouse Lives
    [Documentation]    The Stock location is what makes a warehouse able to hold anything.
    ...    Archiving it out from under a live warehouse would leave the warehouse unusable
    ...    with nothing saying why, so the warehouse is archived instead.
    [Tags]    negative
    Ensure Warehouse Under Test
    ${stock}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Stock
    Should Not Be Empty    ${stock}

    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${stock}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${stock}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A System Generated Location Cannot Be Moved
    [Documentation]    Its place in the tree is what the warehouse flow created it for.
    [Tags]    negative
    ${stock}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Stock
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${stock}/move
    ...    json=${{ {'parent_location_id': $LIFECYCLE_LOCATION_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Client Cannot Mint A System Generated Location
    [Documentation]    The flag is what protects a location from being restructured or
    ...    archived, so a client able to set it could create one nobody can clean up. Create
    ...    strips it rather than refusing, because the rest of the request is perfectly valid.
    ${name}=    Unique Display Name    Robot Pretender
    ${code}=    Unique Code    pretend
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'is_system_generated': True, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Not Be True    ${item}[is_system_generated]
    ...    msg=Only the warehouse service may mark a location system-generated
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${id}    expected_status=any

There Is No Activate Or Deactivate Action
    [Tags]    negative
    FOR    ${action}    IN    activate    deactivate
        ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}/${action}
        ...    json=${{ {} }}    expected_status=any
        Should Be True    ${resp.status_code} >= 400
        ...    msg=/${action} must not be served
    END


*** Keywords ***
Ensure Lifecycle Location
    [Documentation]    A location of its own for this suite, so the states it walks through do
    ...    not disturb the location under test that the CRUD suites share.
    ${id}=    Get Variable Value    ${LIFECYCLE_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    ${name}=    Unique Display Name    Robot Lifecycle Location
    ${code}=    Unique Code    lifeloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${LIFECYCLE_LOCATION_ID}    ${id}

Get Lifecycle Location
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${LIFECYCLE_LOCATION_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    RETURN    ${item}

Create Child Of Lifecycle Location
    ${name}=    Unique Display Name    Robot Lifecycle Child
    ${code}=    Unique Code    lifekid
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'parent_location_id': $LIFECYCLE_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}
