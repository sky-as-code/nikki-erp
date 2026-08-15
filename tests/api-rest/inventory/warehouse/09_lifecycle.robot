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
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/suspend    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[status]    suspended
    Should Not Be True    ${item}[is_archived]
    ...    msg=Suspending must not archive: the warehouse is closed, not withdrawn

Suspending A Suspended Warehouse Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/suspend
    ...    json=${{ {} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Suspending Does Not Cascade To The Warehouse Locations
    [Documentation]    A warehouse can own thousands of locations, so suspending it does not
    ...    rewrite them all. Usability is read from both records instead: a location in a
    ...    suspended warehouse is unusable however the location itself reads.
    ${stock}=    Find Warehouse Location By Code    ${WAREHOUSE_ID}    Stock
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${stock}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    Should Be Equal    ${item}[status]    active
    ...    msg=The location keeps its own state; the warehouse's is read alongside it

Resume Returns The Warehouse To Service
    [Documentation]    TS-STATUS-02.
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/resume    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[status]    active

Resuming An Active Warehouse Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/resume
    ...    json=${{ {} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Unarchive Leaves The Warehouse Suspended
    [Documentation]    TS-STATUS-03 and INV-STATUS-011. Unarchiving never puts a warehouse
    ...    straight back to work: what it sat in may have changed while it was away, so
    ...    someone confirms the configuration through Resume.
    ${item}=    Get Warehouse Under Test
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Warehouse Under Test
    Should Be True    ${item}[is_archived]
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/archived
    ...    json=${{ {'is_archived': False, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Warehouse Under Test
    Should Not Be True    ${item}[is_archived]
    Should Be Equal    ${item}[status]    suspended
    ...    msg=Unarchiving returns a warehouse to suspended, never straight to active

    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/resume    json=${{ {} }}
    Response Status Should Be    ${resp}    200

There Is No Activate Or Deactivate Action
    [Documentation]    The change request removed both verbs. Suspension is the reversible
    ...    state and archiving is the withdrawal; a third pair meaning one of those would only
    ...    drift apart from it.
    [Tags]    negative
    FOR    ${action}    IN    activate    deactivate
        ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/${action}
        ...    json=${{ {} }}    expected_status=any
        Should Be True    ${resp.status_code} >= 400
        ...    msg=/${action} must not be served
    END


*** Keywords ***
Get Warehouse Under Test
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    Set Global Variable    ${WAREHOUSE_ETAG}    ${item}[etag]
    RETURN    ${item}
