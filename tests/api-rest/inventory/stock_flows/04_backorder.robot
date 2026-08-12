*** Settings ***
Documentation     Partial processing and backorder (BR §4.2.3.11, AC-STOCK-008, STOCK-INV-010,
...               STOCK-INV-020).
...
...               The rule that shapes every assertion here: the original transfer's demand is
...               never rewritten to what was actually delivered. A transfer that asked for 100
...               and shipped 70 must remain a transfer that asked for 100 and shipped 70 — that
...               is precisely the question a backorder exists to answer.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Stock Destination Location Under Test
...               AND    Ensure Internal Operation Type Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Seed Stock For Backorder
Test Tags         inventory    stock_flows    backorder


*** Test Cases ***
Partial Processing Leaves The Original Demand Intact
    [Documentation]    STOCK-INV-020 and AC-STOCK-008. Rewriting the demand to the processed
    ...    quantity would erase the fact that more was promised than delivered.
    ${available}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${demand}=    Evaluate    int(${available}) + 50

    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    ${demand}
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${check}=    GET On Session    api    ${STOCK_MOVE_API}/${move_id}
    ${actual}=    Convert To Number    ${check.json()}[data][demand_quantity]
    Should Be Equal As Numbers    ${actual}    ${demand}
    ...    msg=A partial validate must not rewrite the original demand

    Set Suite Variable    ${PARTIAL_TRANSFER_ID}    ${id}
    Set Suite Variable    ${PARTIAL_MOVE_ID}    ${move_id}
    Set Suite Variable    ${PARTIAL_DEMAND}    ${demand}
    Set Suite Variable    ${PARTIAL_PROCESSED}    ${available}

Partial Processing Creates A Backorder
    [Documentation]    The fixture type's policy is 'always', so the undelivered remainder becomes
    ...    a new transfer rather than being silently dropped.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${backorders}=    Evaluate
    ...    [i for i in $resp.json()['data']['items'] if i.get('backorder_of_id') == $PARTIAL_TRANSFER_ID]
    Should Not Be Empty    ${backorders}
    ...    msg=An 'always' backorder policy must raise a new transfer for the remainder
    Set Suite Variable    ${BACKORDER_TRANSFER_ID}    ${backorders}[0][id]

Backorder Points Back At Its Original
    [Documentation]    STOCK-INV-010. The link is what makes a split delivery traceable in both
    ...    directions after the fact.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${BACKORDER_TRANSFER_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[data][backorder_of_id]    ${PARTIAL_TRANSFER_ID}

Backorder Starts As A Fresh Draft
    [Documentation]    It is a new document with its own number and its own lifecycle, not a
    ...    reopening of the original — which stays Done.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${BACKORDER_TRANSFER_ID}
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    draft
    Should Not Be Empty    ${item}[transfer_number]

    ${original}=    GET On Session    api    ${STOCK_TRANSFER_API}/${PARTIAL_TRANSFER_ID}
    Should Be Equal    ${original.json()}[data][status]    done
    ...    msg=The original transfer stays done: what it delivered, it delivered

Backorder Carries Only The Undelivered Quantity
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${moves}=    Evaluate
    ...    [i for i in $resp.json()['data']['items'] if i.get('transfer_id') == $BACKORDER_TRANSFER_ID]
    Should Not Be Empty    ${moves}    msg=The backorder must carry a move for the remainder
    ${carried}=    Convert To Number    ${moves}[0][demand_quantity]
    ${expected}=    Evaluate    ${PARTIAL_DEMAND} - ${PARTIAL_PROCESSED}
    Should Be Equal As Numbers    ${carried}    ${expected}
    ...    msg=The backorder carries exactly what the original did not deliver

Ask Policy Requires An Explicit Decision
    [Documentation]    BR §4.2.3.11. Defaulting either way would make the setting meaningless: one
    ...    default silently drops a commitment to the customer, the other silently creates
    ...    paperwork nobody asked for.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Ask Type
    ${code}=    Unique Code    stkask
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'ask', 'shipping_policy': 'partial', 'org_id': $INV_ORG_ID} }}
    ${type_id}    ${type_etag}=    Response Should Be Create Success    ${resp}

    ${available}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${demand}=    Evaluate    int(${available}) + 50
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $type_id, 'source_location_id': $STOCK_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    ${demand}
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/confirm    json=${{ {} }}    expected_status=any
    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/reserve    json=${{ {} }}    expected_status=any

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=An 'ask' policy with an undelivered remainder must require create_backorder

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${id}/validate
    ...    json=${{ {'create_backorder': False} }}    expected_status=any
    #    An explicit decision must be accepted, whichever way it goes.
    Response Status Should Be    ${resp}    200


*** Keywords ***
Seed Stock For Backorder
    [Documentation]    Leaves a modest, known quantity in the source location, so that a demand
    ...    larger than it is guaranteed to be only partly deliverable.
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    IF    ${on_hand} >= 20    RETURN
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}    40
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
