*** Settings ***
Documentation     Suspend, resume and unarchive. These are the operations that make status a
...               separate concept from archiving: suspension is a temporary, reversible close,
...               archiving is withdrawal from the working set, and neither implies the other.
...               Runs after the CRUD suites so the warehouse under test already exists.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Warehouse Under Test
Test Tags         inventory    warehouse    lifecycle


*** Test Cases ***
Suspend Closes The Warehouse Temporarily
    [Documentation]    TS-STATUS-01. A warehouse shut for a stocktake or a repair is suspended,
    ...    not archived: it is still part of the operational structure and is expected back.
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/suspend    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[status]    suspended
    Should Not Be True    ${item}[is_archived]
    ...    msg=Suspending must not archive: the warehouse is closed, not withdrawn

Suspending A Suspended Warehouse Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/suspend
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Suspending Does Not Cascade To The Warehouse Locations
    [Documentation]    A warehouse can own thousands of locations, so suspending it does not
    ...    rewrite them all. Usability is read from both records instead: a location in a
    ...    suspended warehouse is unusable however the location itself reads.
    ${stock}=    Find Warehouse Location    ${WAREHOUSE_ID}    Stock
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${stock}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[status]    active
    ...    msg=The location keeps its own state; the warehouse's is read alongside it

Resume Returns The Warehouse To Service
    [Documentation]    TS-STATUS-02.
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/resume    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[status]    active

Resuming An Active Warehouse Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/resume
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Unarchive Leaves The Warehouse Suspended
    [Documentation]    TS-STATUS-03 and INV-STATUS-011. Unarchiving never puts a warehouse
    ...    straight back to work: what it sat in may have changed while it was away, so
    ...    someone confirms the configuration through Resume.
    ...    Runs on a warehouse of its own: the root's archive guard would refuse while other
    ...    suites keep locations under the warehouse under test.
    ${id}    ${root}=    Create Lifecycle Warehouse    U
    Set Warehouse Archived    ${id}    ${True}
    ${item}=    Get Warehouse By Id    ${id}
    Should Be True    ${item}[is_archived]
    Set Warehouse Archived    ${id}    ${False}

    ${item}=    Get Warehouse By Id    ${id}
    Should Not Be True    ${item}[is_archived]
    Should Be Equal    ${item}[status]    suspended
    ...    msg=Unarchiving returns a warehouse to suspended, never straight to active

    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${id}/resume    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200

Archive And Unarchive Carry The Root And Stock Locations
    [Documentation]    INV-009. The locations a warehouse manages live and go with it: archiving
    ...    the warehouse archives its Stock and root in the same transaction, and unarchiving
    ...    brings them back.
    ${id}    ${root}=    Create Lifecycle Warehouse    A
    ${stock}=    Get Warehouse Stock Location Id    ${id}
    Set Warehouse Archived    ${id}    ${True}
    FOR    ${location_id}    IN    ${root}    ${stock}
        ${location}=    Get Location    ${location_id}
        Should Be True    ${location}[is_archived]    msg=${location}[name] is archived with its warehouse
    END
    Set Warehouse Archived    ${id}    ${False}
    FOR    ${location_id}    IN    ${root}    ${stock}
        ${location}=    Get Location    ${location_id}
        Should Not Be True    ${location}[is_archived]    msg=${location}[name] is unarchived with its warehouse
    END

Archiving A Warehouse Archives Every Location It Manages
    [Documentation]    The flow stops are archived with the warehouse too (stops, then Stock, then
    ...    the root), so only a user's own sub-location can keep a warehouse from being archived.
    ...    Unarchiving brings back the stops the current flows use.
    ${id}    ${root}=    Create Lifecycle Warehouse    S    three_step    three_step
    Set Warehouse Archived    ${id}    ${True}
    FOR    ${stop}    IN    Stock    Input    Quality Control    Packing    Output
        ${location_id}=    Find Warehouse Location    ${id}    ${stop}
        ${location}=    Get Location    ${location_id}
        Should Be True    ${location}[is_archived]    msg=${stop} is archived with its warehouse
    END
    ${location}=    Get Location    ${root}
    Should Be True    ${location}[is_archived]
    Set Warehouse Archived    ${id}    ${False}
    FOR    ${stop}    IN    Stock    Input    Quality Control    Packing    Output
        ${location_id}=    Find Warehouse Location    ${id}    ${stop}
        ${location}=    Get Location    ${location_id}
        Should Not Be True    ${location}[is_archived]    msg=${stop} is unarchived with its warehouse
    END
    DELETE On Session    api    ${WAREHOUSE_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Archiving A Warehouse Whose Stock Has A Live Child Is Refused
    [Documentation]    Each managed location's own archive guard applies: a sub-location still in
    ...    use keeps Stock, and so the warehouse, in service. Nothing is archived.
    [Tags]    negative
    ${id}    ${root}=    Create Lifecycle Warehouse    C
    ${stock}=    Get Warehouse Stock Location Id    ${id}
    ${child}=    Create Location Under    ${stock}
    ${item}=    Get Warehouse By Id    ${id}
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${id}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'is_archived': True, 'etag': $item['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ${item}=    Get Warehouse By Id    ${id}
    Should Not Be True    ${item}[is_archived]    msg=A refused Stock archive rolls the warehouse back
    ${location}=    Get Location    ${root}
    Should Not Be True    ${location}[is_archived]    msg=The root is rolled back with it
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${child}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    DELETE On Session    api    ${WAREHOUSE_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Deleting A Warehouse Deletes Every Location It Manages
    [Documentation]    INV-009. Deleting the warehouse releases its RESTRICT FKs; its stops,
    ...    Stock and root are then deleted through DeleteManaged in the same transaction.
    ${id}    ${root}=    Create Lifecycle Warehouse    D    three_step    three_step
    ${managed}=    Create List    ${root}
    FOR    ${stop}    IN    Stock    Input    Quality Control    Packing    Output
        ${location_id}=    Find Warehouse Location    ${id}    ${stop}
        ${managed}=    Evaluate    $managed + [$location_id]
    END
    ${resp}=    DELETE On Session    api    ${WAREHOUSE_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    FOR    ${location_id}    IN    @{managed}
        ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${location_id}
        ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
        Response Should Be Not Found Error    ${resp}
    END

Deleting A Warehouse Whose Stock Has A Child Is Refused
    [Documentation]    Stock cannot go while a location still hangs under it, and the
    ...    warehouse does not go without it: the whole delete rolls back.
    [Tags]    negative
    ${id}    ${root}=    Create Lifecycle Warehouse    E
    ${stock}=    Get Warehouse Stock Location Id    ${id}
    ${child}=    Create Location Under    ${stock}
    ${resp}=    DELETE On Session    api    ${WAREHOUSE_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${child}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    DELETE On Session    api    ${WAREHOUSE_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

The Root Location Cannot Be Deleted On Its Own
    [Documentation]    A managed record is deleted by its manager only.
    [Tags]    negative
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${root}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    Should Contain    ${resp.text}    err_record_managed_delete

The Root Location Cannot Be Archived On Its Own
    [Tags]    negative
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${location}=    Get Location    ${root}
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${root}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'is_archived': True, 'etag': $location['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    Should Contain    ${resp.text}    inventory_location.archive_with_warehouse

The Root Location Type Is Fixed But Its Name Is Not
    [Documentation]    INV-009/010. The root is the group every location of the warehouse hangs
    ...    under and holds no stock, so its type cannot change; renaming it is fine.
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${location}=    Get Location    ${root}
    ${resp}=    PATCH On Session    api    ${INVENTORY_LOCATION_API}/${root}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'location_type': 'internal', 'etag': $location['etag']} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    Should Contain    ${resp.text}    inventory_location.root_type_immutable
    ${resp}=    PATCH On Session    api    ${INVENTORY_LOCATION_API}/${root}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': 'Robot Root', 'etag': $location['etag']} }}
    Response Should Be Update Success    ${resp}
    ${location}=    Get Location    ${root}
    Should Be Equal    ${location}[canonical_name]    Robot Root

There Is No Activate Or Deactivate Action
    [Documentation]    The change request removed both verbs. Suspension is the reversible
    ...    state and archiving is the withdrawal; a third pair meaning one of those would only
    ...    drift apart from it.
    [Tags]    negative
    FOR    ${action}    IN    activate    deactivate
        ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/${action}
        ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
        Should Be True    ${resp.status_code} >= 400
        ...    msg=/${action} must not be served
    END


*** Keywords ***
Get Warehouse Under Test
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    Set Global Variable    ${WAREHOUSE_ETAG}    ${item}[etag]
    RETURN    ${item}

Get Warehouse By Id
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    RETURN    ${item}

Get Location
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    RETURN    ${item}

Create Lifecycle Warehouse
    [Documentation]    A warehouse of the test's own (one-step by default, so it has no stops),
    ...    so the archive and delete cascades are not blocked by what other suites hung under the
    ...    warehouse under test. Returns the warehouse and root ids.
    [Arguments]    ${prefix}    ${incoming}=one_step    ${outgoing}=one_step
    ${name}=    Unique Display Name    Robot Lifecycle Warehouse
    ${code}=    Unique Warehouse Code    ${prefix}
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': $name, 'warehouse_role': 'other', 'incoming_flow': $incoming, 'outgoing_flow': $outgoing, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${root}=    Get Warehouse Root Location Id    ${id}
    RETURN    ${id}    ${root}

Set Warehouse Archived
    [Arguments]    ${id}    ${archived}
    ${item}=    Get Warehouse By Id    ${id}
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${id}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'is_archived': $archived, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

Create Location Under
    [Arguments]    ${parent_id}
    ${name}=    Unique Display Name    Robot Child
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'name': $name, 'location_type': 'internal', 'parent_location_id': $parent_id, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}
