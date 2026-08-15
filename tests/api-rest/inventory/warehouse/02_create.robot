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
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'central', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${WAREHOUSE_ID}    ${id}
    Set Global Variable    ${WAREHOUSE_ETAG}    ${etag}
    Set Global Variable    ${WAREHOUSE_CODE}    ${code}

Creating A Warehouse Creates Its Root And Stock Locations
    [Documentation]    TS-03. A warehouse with no Stock location could hold nothing, so the two
    ...    are created together. The root is named after the warehouse code, which is what
    ...    makes a path read 'MAIN/Stock'.
    ${root}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    ${WAREHOUSE_CODE}
    Should Not Be Empty    ${root}    msg=A warehouse must have a root location
    ${stock}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Stock
    Should Not Be Empty    ${stock}    msg=A warehouse must have a Stock location

    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${stock}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[warehouse_id]    ${WAREHOUSE_ID}
    Should Be Equal    ${item}[location_usage]    internal
    Should Be Equal    ${item}[purpose]    storage
    Should Be True    ${item}[is_system_generated]
    ...    msg=Locations the warehouse created for itself must be marked as such
    Should Be Equal    ${item}[complete_path]    ${WAREHOUSE_CODE}/Stock

A One Step Warehouse Has No Input Or Output Location
    [Documentation]    The stops exist only when a flow asks for them. Creating them anyway
    ...    would advertise places goods never pass through.
    ${input}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Input
    Should Be Empty    ${input}
    ${output}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Output
    Should Be Empty    ${output}

Create Three Step Warehouse Creates Every Stop
    [Documentation]    A three-step flow in each direction needs Input and Quality Control on
    ...    the way in, and Packing and Output on the way out. All five locations plus the root
    ...    arrive with the warehouse, in one transaction.
    ${name}=    Unique Display Name    Robot Three Step
    ${code}=    Unique Warehouse Code    T
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'central', 'incoming_flow': 'three_step', 'outgoing_flow': 'three_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}

    FOR    ${expected}    IN    Stock    Input    Quality Control    Packing    Output
        ${found}=    Find Warehouse Location By Code    ${id}    ${expected}
        Should Not Be Empty    ${found}    msg=A three-step warehouse needs a '${expected}' location
    END
    Archive Warehouse Fixture By Id    ${id}

Create With Parent Warehouse Succeeds
    [Documentation]    Warehouses form an organisational hierarchy. It says which warehouse
    ...    system a site belongs to and nothing about where stock is: a POS under a central
    ...    warehouse holds its own goods.
    Ensure Secondary Warehouse Under Test
    ${resp}=    PATCH On Session    api    ${WAREHOUSE_API}/${SECONDARY_WAREHOUSE_ID}
    ...    json=${{ {'parent_warehouse_id': $WAREHOUSE_ID, 'etag': $SECONDARY_WAREHOUSE_ETAG} }}
    Response Should Be Update Success    ${resp}
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${SECONDARY_WAREHOUSE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    Should Be Equal    ${item}[parent_warehouse_id]    ${WAREHOUSE_ID}
    Set Global Variable    ${SECONDARY_WAREHOUSE_ETAG}    ${item}[etag]

Create With Duplicate Code Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $WAREHOUSE_CODE, 'name': {'en-US': $name}, 'warehouse_role': 'other', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}

Create With Unknown Flow Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Flow
    ${code}=    Unique Warehouse Code    B
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'other', 'incoming_flow': 'four_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A flow outside the declared enum must not be accepted


*** Keywords ***
Archive Warehouse Fixture By Id
    [Documentation]    Withdraws a warehouse created inside one test. It is archived rather
    ...    than deleted because it owns the locations it created, and those refuse to go while
    ...    it is live.
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${id}    expected_status=any
    IF    ${resp.status_code} == 200
        POST On Session    api    ${WAREHOUSE_API}/${id}/archived
        ...    json=${{ {'is_archived': True, 'etag': $resp.json()['etag']} }}    expected_status=any
    END
