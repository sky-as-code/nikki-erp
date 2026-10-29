*** Settings ***
Documentation     Creating Warehouses. The first test saves the warehouse under test
...               (${WAREHOUSE_ID}/${WAREHOUSE_ETAG}) consumed by the later suites.
...               Creating a warehouse also creates the locations it needs to function, which
...               is asserted here rather than in the location suite: the warehouse is what
...               makes them exist.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    warehouse    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Warehouse
    ${code}=    Unique Warehouse Code    W
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': $name, 'warehouse_role': 'central', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${WAREHOUSE_ID}    ${id}
    Set Global Variable    ${WAREHOUSE_ETAG}    ${etag}
    Set Global Variable    ${WAREHOUSE_CODE}    ${code}

The Root Location Is A Group Named After The Warehouse Code
    [Documentation]    TS-03 / INV-009. The root is created with the warehouse, in one
    ...    transaction: named after the warehouse code, managed by the warehouse, and of type
    ...    group, so it organises the warehouse's locations and holds no stock itself.
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    Should Not Be Empty    ${root}    msg=A warehouse must have a root location

    ${item}=    Get Location By Id    ${root}
    Should Be Equal    ${item}[name]    ${WAREHOUSE_CODE}
    Should Be Equal    ${item}[canonical_name]    ${WAREHOUSE_CODE}
    Should Be Equal    ${item}[warehouse_id]    ${WAREHOUSE_ID}
    Should Be Equal    ${item}[location_type]    group
    Dictionary Should Not Contain Key    ${item}    purpose
    Should Be Equal    ${item}[managed_by_type]    commerce/inventory_warehouse
    Should Be Equal    ${item}[managed_by_id]    ${WAREHOUSE_ID}
    Should Be True    ${item}[is_system_generated]
    ...    msg=Locations the warehouse created for itself must be marked as such

Creating A Warehouse Creates Its Stock Location
    [Documentation]    A warehouse with nowhere to hold stock could hold nothing, so an internal
    ...    Stock location is created under the root with it and referenced by stock_location_id.
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${stock}=    Get Warehouse Stock Location Id    ${WAREHOUSE_ID}
    Should Not Be Empty    ${stock}    msg=A warehouse must have a Stock location

    ${item}=    Get Location By Id    ${stock}
    Should Be Equal    ${item}[name]    Stock
    Should Be Equal    ${item}[canonical_name]    ${WAREHOUSE_CODE} / Stock
    Should Be Equal    ${item}[parent_location_id]    ${root}
    Should Be Equal    ${item}[warehouse_id]    ${WAREHOUSE_ID}
    Should Be Equal    ${item}[location_type]    internal
    Should Be Equal    ${item}[managed_by_type]    commerce/inventory_warehouse
    Should Be Equal    ${item}[managed_by_id]    ${WAREHOUSE_ID}
    Should Be True    ${item}[is_system_generated]

A Move Into The Group Root Is Refused
    [Documentation]    A group location holds no stock: a transfer naming the warehouse's root as
    ...    its destination is refused, and goods go to Stock (or a location inside it) instead.
    [Tags]    negative
    Ensure Supplier Location Under Test
    Ensure Receipt Operation Type Under Test
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $RECEIPT_OPERATION_TYPE_ID, 'source_location_id': $SUPPLIER_LOCATION_ID, 'destination_location_id': $root, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    Should Contain    ${resp.text}    inventory_location.group_holds_no_stock

The Root Location Id Cannot Be Changed By An Update
    [Documentation]    root_location_id is auto-generated: an update carrying another id is
    ...    accepted for its other fields and leaves the root alone.
    Ensure Inventory Location Under Test
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    ${resp}=    PATCH On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'root_location_id': $INVENTORY_LOCATION_ID, 'notes': 'root stays', 'etag': $item['etag']} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    IF    $etag is not None    Set Global Variable    ${WAREHOUSE_ETAG}    ${etag}
    ${after}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    Should Be Equal    ${after}    ${root}

A One Step Warehouse Has No Input Or Output Location
    [Documentation]    The stops exist only when a flow asks for them. Creating them anyway
    ...    would advertise places goods never pass through.
    ${input}=    Find Warehouse Location    ${WAREHOUSE_ID}    Input
    Should Be Empty    ${input}
    ${output}=    Find Warehouse Location    ${WAREHOUSE_ID}    Output
    Should Be Empty    ${output}

Create Three Step Warehouse Creates Every Stop
    [Documentation]    A three-step flow in each direction needs Input and Quality Control on
    ...    the way in, and Packing and Output on the way out. All five locations plus the root
    ...    arrive with the warehouse, in one transaction.
    ${name}=    Unique Display Name    Robot Three Step
    ${code}=    Unique Warehouse Code    T
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': $name, 'warehouse_role': 'central', 'incoming_flow': 'three_step', 'outgoing_flow': 'three_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${root}=    Get Warehouse Root Location Id    ${id}
    FOR    ${expected}    IN    Stock    Input    Quality Control    Packing    Output
        ${found}=    Find Warehouse Location    ${id}    ${expected}
        Should Not Be Empty    ${found}    msg=A three-step warehouse needs a '${expected}' location
        ${item}=    Get Location By Id    ${found}
        Should Be Equal    ${item}[parent_location_id]    ${root}
        Should Be Equal    ${item}[managed_by_id]    ${id}
    END
    Archive Warehouse Fixture By Id    ${id}


Create With Duplicate Code Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $WAREHOUSE_CODE, 'name': $name, 'warehouse_role': 'other', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Unknown Flow Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Flow
    ${code}=    Unique Warehouse Code    B
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': $name, 'warehouse_role': 'other', 'incoming_flow': 'four_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A flow outside the declared enum must not be accepted


*** Keywords ***
Archive Warehouse Fixture By Id
    [Documentation]    Withdraws a warehouse created inside one test. It is archived rather
    ...    than deleted because it owns the locations it created, and those refuse to go while
    ...    it is live.
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    IF    ${resp.status_code} == 200
        POST On Session    api    ${WAREHOUSE_API}/${id}/archived
        ...    json=${{ {'org_id': $INV_ORG_ID, 'is_archived': True, 'etag': $resp.json()['item']['etag']} }}    expected_status=any
    END
