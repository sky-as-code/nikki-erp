*** Settings ***
Documentation     Creating putaway rules and asking them where goods should go.
...
...               A rule answers a question and nothing more: the suggestion carries a
...               destination and the rule that produced it, and no quantity moves. Acting on
...               the answer is the Stock movement engine's job.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
...               AND    Ensure Warehouse Under Test    AND    Ensure Putaway Locations
Test Tags         inventory    putaway_rule    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${code}=    Unique Code    putaway
    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}
    ...    json=${{ {'code': $code, 'warehouse_id': $WAREHOUSE_ID, 'source_location_id': $PUTAWAY_ARRIVAL_ID, 'destination_location_id': $PUTAWAY_DEST_ID, 'priority': 5, 'sublocation_strategy': 'fixed', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PUTAWAY_RULE_ID}    ${id}
    Set Global Variable    ${PUTAWAY_RULE_ETAG}    ${etag}

Suggest Returns The Matching Destination
    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $PUTAWAY_ARRIVAL_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[destination_location_id]    ${PUTAWAY_DEST_ID}
    Should Be Equal    ${resp.json()}[matched_rule_id]    ${PUTAWAY_RULE_ID}

Suggest Changes No Stock
    [Documentation]    AC-CR-LOC-032. Asking where goods should go is a read. If the lookup
    ...    moved anything, a caller could not ask twice without consequences.
    ${before}=    Count Transfers In Org
    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $PUTAWAY_ARRIVAL_ID} }}
    ${after}=    Count Transfers In Org
    Should Be Equal As Integers    ${before}    ${after}

The Lowest Priority Rule Wins
    [Documentation]    Several rules may match one arrival. Priority decides, lowest first, so
    ...    a specific rule can be placed ahead of a general one.
    ${second_dest}=    Create Putaway Location    putdest2    internal
    ${code}=    Unique Code    putfirst
    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}
    ...    json=${{ {'code': $code, 'warehouse_id': $WAREHOUSE_ID, 'source_location_id': $PUTAWAY_ARRIVAL_ID, 'destination_location_id': $second_dest, 'priority': 1, 'sublocation_strategy': 'fixed', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $PUTAWAY_ARRIVAL_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[destination_location_id]    ${second_dest}
    ...    msg=Priority 1 must be considered before priority 5

    DELETE On Session    api    ${PUTAWAY_RULE_API}/${id}    expected_status=any

A Suspended Destination Is Skipped
    [Documentation]    TS-STATUS-05. A suspended location still exists and still holds whatever
    ...    it held, but nothing new is routed to it. Suggesting somewhere goods may not go
    ...    would be worse than suggesting nowhere.
    POST On Session    api    ${INVENTORY_LOCATION_API}/${PUTAWAY_DEST_ID}/suspend    json=${{ {} }}

    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $PUTAWAY_ARRIVAL_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[destination_location_id]
    ...    msg=A rule pointing at a suspended location suggests nothing

    POST On Session    api    ${INVENTORY_LOCATION_API}/${PUTAWAY_DEST_ID}/resume    json=${{ {} }}

An Archived Rule Is Not Evaluated
    [Documentation]    TS-STATUS-09. Archiving is the whole of a rule's lifecycle: an archived
    ...    rule is out of the working set, so it takes no part in the decision.
    ${resp}=    GET On Session    api    ${PUTAWAY_RULE_API}/${PUTAWAY_RULE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/putaway_rule.json    200
    POST On Session    api    ${PUTAWAY_RULE_API}/${PUTAWAY_RULE_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}

    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $PUTAWAY_ARRIVAL_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[destination_location_id]

    ${resp}=    GET On Session    api    ${PUTAWAY_RULE_API}/${PUTAWAY_RULE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/putaway_rule.json    200
    POST On Session    api    ${PUTAWAY_RULE_API}/${PUTAWAY_RULE_ID}/archived
    ...    json=${{ {'is_archived': False, 'etag': $item['etag']} }}

Suggest With No Matching Rule Answers Nothing
    [Documentation]    Not an error: the caller falls back to whatever default it had in mind.
    ${elsewhere}=    Create Putaway Location    putelse    internal
    ${resp}=    POST On Session    api    ${PUTAWAY_RULE_API}/suggest_location
    ...    json=${{ {'warehouse_id': $WAREHOUSE_ID, 'arrival_location_id': $elsewhere} }}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[destination_location_id]
    Should Be Empty    ${resp.json()}[matched_rule_id]


*** Keywords ***
Ensure Putaway Locations
    [Documentation]    An arrival location and somewhere for goods to be put, both inside the
    ...    warehouse under test: a rule may not route goods between warehouses.
    ${id}=    Get Variable Value    ${PUTAWAY_ARRIVAL_ID}    ${EMPTY}
    IF    $id    RETURN
    ${arrival}=    Create Putaway Location    putarr    internal
    Set Global Variable    ${PUTAWAY_ARRIVAL_ID}    ${arrival}
    ${dest}=    Create Putaway Location    putdest    internal
    Set Global Variable    ${PUTAWAY_DEST_ID}    ${dest}

Create Putaway Location
    [Arguments]    ${prefix}    ${usage}
    ${name}=    Unique Display Name    Robot ${prefix}
    ${code}=    Unique Code    ${prefix}
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': $usage, 'warehouse_id': $WAREHOUSE_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Count Transfers In Org
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[total]
