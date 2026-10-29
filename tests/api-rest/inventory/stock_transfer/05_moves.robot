*** Settings ***
Documentation     A transfer's moves written with the transfer: create and update carry a
...               `moves: {add, update, remove}` delta, the same shape as the IAM role's
...               `entitlements`. The transfer row is written first and the moves after it, in one
...               transaction, so a refused move rolls the whole request back. Moves change only
...               while the transfer is draft.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
Test Tags         inventory    stock_transfer    moves


*** Test Cases ***
Create With Moves Creates The Moves In Sequence
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '3'}, {'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '4.5'}] }}
    ${moves}=    Transfer Moves    ${id}
    Length Should Be    ${moves}    2
    Should Be Equal As Integers    ${moves}[0][sequence]    1
    Should Be Equal As Integers    ${moves}[1][sequence]    2
    Should Be Equal As Numbers    ${moves}[0][demand_quantity]    3
    Should Be Equal As Numbers    ${moves}[1][base_demand_quantity]    4.5
    Should Be Equal    ${moves}[0][status]    draft
    Should Be Equal    ${moves}[0][source_location_id]    ${INVENTORY_LOCATION_ID}
    Should Be Equal    ${moves}[0][destination_location_id]    ${STOCK_DEST_LOCATION_ID}
    ...    msg=A move takes its locations from the transfer
    [Teardown]    Delete Transfer With Moves    ${id}

Get Transfer Returns Its Moves Through The Moves Edge
    [Documentation]    `moves` is a one-to-many edge of the transfer, so a detail page reads the
    ...    header and its moves in one request.
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '2'}] }}
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': 'id,moves.id,moves.demand_quantity,moves.sequence'} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[item][moves]    1
    Should Be Equal As Numbers    ${resp.json()}[item][moves][0][demand_quantity]    2
    [Teardown]    Delete Transfer With Moves    ${id}

Update Adds, Updates And Removes Moves
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}, {'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '2'}] }}
    ${moves}=    Transfer Moves    ${id}
    ${first}=    Set Variable    ${moves}[0]
    ${second}=    Set Variable    ${moves}[1]
    ${delta}=    Set Variable    ${{ {'add': [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '9'}], 'update': [{'id': $first['id'], 'etag': $first['etag'], 'demand_quantity': '7'}], 'remove': [$second['id']]} }}
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${id}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $etag, 'moves': $delta} }}
    Response Should Be Update Success    ${resp}    previous_etag=${etag}

    ${after}=    Transfer Moves    ${id}
    Length Should Be    ${after}    2
    Should Be Equal    ${after}[0][id]    ${first}[id]
    Should Be Equal As Numbers    ${after}[0][demand_quantity]    7
    Should Be Equal As Numbers    ${after}[0][base_demand_quantity]    7
    Should Be Equal As Integers    ${after}[1][sequence]    3
    ...    msg=An added move continues the sequence after the moves the transfer had
    Should Be Equal As Numbers    ${after}[1][demand_quantity]    9
    [Teardown]    Delete Transfer With Moves    ${id}

Update Replaces A Move Whose Variant Changes
    [Documentation]    A move's variant is fixed once created, so a changed variant replaces the
    ...    draft move in the same position.
    ${key}=    Unique Code    comb
    ${other_variant}    ${other_etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '5'}] }}
    ${moves}=    Transfer Moves    ${id}
    ${move}=    Set Variable    ${moves}[0]
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${id}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $etag, 'moves': {'update': [{'id': $move['id'], 'etag': $move['etag'], 'product_variant_id': $other_variant}]}} }}
    Response Should Be Update Success    ${resp}
    ${after}=    Transfer Moves    ${id}
    Length Should Be    ${after}    1
    Should Not Be Equal    ${after}[0][id]    ${move}[id]
    Should Be Equal    ${after}[0][product_variant_id]    ${other_variant}
    Should Be Equal As Integers    ${after}[0][sequence]    ${move}[sequence]
    Should Be Equal As Numbers    ${after}[0][demand_quantity]    5
    [Teardown]    Run Keywords    Delete Transfer With Moves    ${id}
    ...    AND    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${other_variant}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Update Of A Move Of Another Transfer Is Refused
    [Tags]    negative
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}] }}
    ${other}    ${other_etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}] }}
    ${foreign}=    Transfer Moves    ${other}
    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${id}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $etag, 'moves': {'remove': [$foreign[0]['id']]}} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:stock_transfer.move_not_in_transfer
    ...    field=moves.remove[0]
    ${kept}=    Transfer Moves    ${other}
    Length Should Be    ${kept}    1
    [Teardown]    Run Keywords    Delete Transfer With Moves    ${id}
    ...    AND    Delete Transfer With Moves    ${other}

Changing Moves After Confirm Is Refused
    [Tags]    negative
    ${id}    ${etag}=    Create Transfer With Moves    ${{ [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}] }}
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${check}=    GET On Session    api    ${STOCK_TRANSFER_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}
    ${etag}=    Set Variable    ${check.json()}[item][etag]

    ${resp}=    PATCH On Session    api    ${STOCK_TRANSFER_API}/${id}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $etag, 'moves': {'add': [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}]}} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:stock_transfer.moves_not_editable    field=moves
    ${moves}=    Transfer Moves    ${id}
    Length Should Be    ${moves}    1
    ...    msg=A refused delta must leave the moves as they were
    [Teardown]    Run Keywords
    ...    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/cancel    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    ...    AND    Delete Transfer With Moves    ${id}

An Invalid Move Rolls Back The Header
    [Documentation]    The header and its moves are one write: a refused move leaves no empty
    ...    transfer behind.
    [Tags]    negative
    ${origin}=    Unique Code    rollback
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'origin_reference': $origin, 'org_id': $INV_ORG_ID, 'moves': {'add': [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '1'}, {'product_variant_id': $NOT_FOUND_ID, 'demand_quantity': '1'}]}} }}
    ...    expected_status=any
    Response Status Should Be    ${resp}    400
    ${fields}=    Evaluate    [e.get('field') or '' for e in $resp.json()]
    Should Be True    any(f.startswith('moves.add[1]') for f in $fields)
    ...    msg=The error must point at the refused move, got: ${resp.text}
    ${search}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if": ["origin_reference", "=", "%s"]}' % $origin} }}
    Response Status Should Be    ${search}    200
    Should Be Empty    ${search.json()}[items]
    ...    msg=A refused move must roll the transfer header back

Create With A Non Positive Quantity Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID, 'moves': {'add': [{'product_variant_id': $PRODUCT_VARIANT_ID, 'demand_quantity': '0'}]}} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:stock_move.quantity_must_be_positive
    ...    field=moves.add[0].demand_quantity

A Malformed Moves Delta Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID, 'moves': ['not', 'a', 'delta']} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    common:err_malformed_request    field=moves


*** Keywords ***
Create Transfer With Moves
    [Arguments]    ${lines}
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $INTERNAL_OPERATION_TYPE_ID, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID, 'moves': {'add': $lines}} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Transfer Moves
    [Documentation]    The transfer's moves, in sequence order.
    [Arguments]    ${transfer_id}
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100, 'fields': 'id,etag,sequence,product_variant_id,demand_quantity,base_demand_quantity,status,source_location_id,destination_location_id', 'graph': '{"if": ["transfer_id", "=", "%s"]}' % $transfer_id} }}
    Response Status Should Be    ${resp}    200
    ${moves}=    Evaluate    sorted($resp.json()['items'], key=lambda m: m['sequence'])
    RETURN    ${moves}

Delete Transfer With Moves
    [Arguments]    ${transfer_id}
    ${moves}=    Transfer Moves    ${transfer_id}
    FOR    ${move}    IN    @{moves}
        DELETE On Session    api    ${STOCK_MOVE_API}/${move}[id]
        ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    END
    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
