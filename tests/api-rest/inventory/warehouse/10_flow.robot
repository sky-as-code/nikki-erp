*** Settings ***
Documentation     Reconfiguring the receipt and delivery flows. A flow is policy: changing it
...               provisions the locations the new shape needs and does nothing else. It
...               creates no stock move, touches no quantity, and leaves a transfer already
...               under way exactly as it was.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Warehouse Under Test
Test Tags         inventory    warehouse    flow


*** Test Cases ***
Widening The Incoming Flow Creates The Stops It Needs
    [Documentation]    TS-09 and AC-CR-LOC-028. Three steps in means goods stop at Input and
    ...    then Quality Control before reaching Stock, so both locations must exist afterwards.
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_incoming_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'three_step'} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[incoming_flow]    three_step

    FOR    ${expected}    IN    Input    Quality Control
        ${found}=    Find Warehouse Location    ${WAREHOUSE_ID}    ${expected}
        Should Not Be Empty    ${found}    msg=Three-step receipt needs a '${expected}' location
    END

Widening The Outgoing Flow Creates The Stops It Needs
    [Documentation]    AC-CR-LOC-029. Three steps out means Packing then Output before the
    ...    goods leave.
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_outgoing_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'three_step'} }}
    Response Status Should Be    ${resp}    200

    FOR    ${expected}    IN    Packing    Output
        ${found}=    Find Warehouse Location    ${WAREHOUSE_ID}    ${expected}
        Should Not Be Empty    ${found}    msg=Three-step delivery needs a '${expected}' location
    END

Narrowing A Flow Archives The Unused Stop And Keeps Its FK
    [Documentation]    Goods that once passed through Quality Control are recorded as having
    ...    done so. Deleting the location would break those records, so it is archived: still
    ...    there, still referenced by quality_location_id, no longer offered for new work.
    ${quality}=    Find Warehouse Location    ${WAREHOUSE_ID}    Quality Control
    Should Not Be Empty    ${quality}

    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_incoming_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'two_step'} }}
    Response Status Should Be    ${resp}    200

    ${warehouse}=    Get Warehouse Under Test
    Should Be Equal    ${warehouse}[quality_location_id]    ${quality}
    ...    msg=The warehouse keeps its FK to a stop the flow no longer uses
    ${item}=    Get Location By Id    ${quality}
    Should Be True    ${item}[is_archived]
    ...    msg=A stop a flow no longer uses is archived, never deleted

    ${input}=    Find Warehouse Location    ${WAREHOUSE_ID}    Input
    ${item}=    Get Location By Id    ${input}
    Should Not Be True    ${item}[is_archived]
    ...    msg=Input is still needed at two steps, so it stays in service

Widening Again Unarchives The Archived Stop
    [Documentation]    Switching a flow back and forth must not accumulate duplicate locations.
    ...    The archived stop the warehouse still references is unarchived rather than a second
    ...    one created.
    ${before}=    Find Warehouse Location    ${WAREHOUSE_ID}    Quality Control

    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_incoming_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'three_step'} }}
    Response Status Should Be    ${resp}    200

    ${after}=    Find Warehouse Location    ${WAREHOUSE_ID}    Quality Control
    Should Be Equal    ${before}    ${after}
    ...    msg=The same location comes back, rather than a duplicate being created

    ${item}=    Get Location By Id    ${after}
    Should Not Be True    ${item}[is_archived]
    Should Be Equal    ${item}[status]    active

Reconfiguring A Flow Creates No Stock Movement
    [Documentation]    TS-09 and AC-CR-LOC-026/027. Warehouse is configuration and Stock is
    ...    execution: changing where goods would go next time moves nothing now.
    ${before}=    Count Transfers In Org

    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_outgoing_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'two_step'} }}
    Response Status Should Be    ${resp}    200

    ${after}=    Count Transfers In Org
    Should Be Equal As Integers    ${before}    ${after}
    ...    msg=A flow change must not create a transfer

Setting A Flow To Its Current Value Changes Nothing
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_outgoing_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'two_step'} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Get Warehouse Under Test
    Should Be Equal    ${item}[outgoing_flow]    two_step

Configuring An Unknown Flow Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}/configure_incoming_flow
    ...    json=${{ {'org_id': $INV_ORG_ID, 'flow': 'four_step'} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400


*** Keywords ***
Get Warehouse Under Test
    ${resp}=    GET On Session    api    ${WAREHOUSE_API}/${WAREHOUSE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/warehouse.json    200
    Set Global Variable    ${WAREHOUSE_ETAG}    ${item}[etag]
    RETURN    ${item}

Count Transfers In Org
    [Documentation]    How many transfers exist, so a test can show an operation created none.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[total]
