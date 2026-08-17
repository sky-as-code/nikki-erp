*** Settings ***
Documentation     Creating supply relations, and the rules about the shape of the resupply
...               graph. A relation only declares who may restock whom: it reserves nothing
...               and starts no transfer, which is asserted here rather than assumed.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
...               AND    Ensure Warehouse Under Test    AND    Ensure Secondary Warehouse Under Test
Test Tags         inventory    supply_relation    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $WAREHOUSE_ID, 'destination_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'priority': 1, 'is_default': True, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SUPPLY_RELATION_ID}    ${id}
    Set Global Variable    ${SUPPLY_RELATION_ETAG}    ${etag}

Creating A Relation Creates No Transfer
    [Documentation]    AC-CR-LOC-035. Declaring that one warehouse may restock another is
    ...    topology. When replenishment actually happens the Stock movement engine creates the
    ...    movement; the declaration alone moves nothing.
    ${before}=    Count Transfers In Org
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'destination_warehouse_id': $WAREHOUSE_ID, 'priority': 9, 'is_default': False, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    # Whether this particular route is allowed is the cycle test's business; what matters here
    # is that no movement appeared either way.
    ${after}=    Count Transfers In Org
    Should Be Equal As Integers    ${before}    ${after}

A Warehouse Cannot Supply Itself
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $WAREHOUSE_ID, 'destination_warehouse_id': $WAREHOUSE_ID, 'priority': 1, 'is_default': False, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400

The Same Route Cannot Be Declared Twice
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $WAREHOUSE_ID, 'destination_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'priority': 5, 'is_default': False, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Destination Has At Most One Default Source
    [Documentation]    A destination may have several sources ranked by priority, but only one
    ...    of them is the default. Two defaults would leave the choice undefined.
    [Tags]    negative
    ${third}=    Create Extra Warehouse    D
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $third, 'destination_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'priority': 2, 'is_default': True, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A second default source for the same destination must be refused

A Second Non Default Source Is Allowed
    [Documentation]    Several warehouses may be able to restock one, ranked by priority. Only
    ...    the default is exclusive.
    ${third}=    Create Extra Warehouse    E
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $third, 'destination_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'priority': 2, 'is_default': False, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${SUPPLY_RELATION_API}/${id}    expected_status=any

A Circular Route Is Refused
    [Documentation]    A supplies B and B supplies A would let replenishment planning chase its
    ...    own tail, with no warehouse actually holding the goods.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SUPPLY_RELATION_API}
    ...    json=${{ {'source_warehouse_id': $SECONDARY_WAREHOUSE_ID, 'destination_warehouse_id': $WAREHOUSE_ID, 'priority': 1, 'is_default': False, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A route back to an upstream warehouse must be refused


*** Keywords ***
Create Extra Warehouse
    [Documentation]    A throwaway warehouse for a rule that needs a third party. Archived by
    ...    the caller's suite teardown rather than deleted, since it owns its own locations.
    [Arguments]    ${prefix}
    ${name}=    Unique Display Name    Robot Extra ${prefix}
    ${code}=    Unique Warehouse Code    ${prefix}
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'other', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Count Transfers In Org
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[total]
