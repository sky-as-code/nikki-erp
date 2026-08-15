*** Settings ***
Documentation     Reverse transfer / return: goods physically coming back (BR §4.2.10).
...
...               Two rules carry this suite. AC-STOCK-021 says a return generates a reverse
...               transfer rather than editing the original, and AC-STOCK-010 with STOCK-INV-005
...               says the original must come through completely untouched — the history has to
...               read original move → reverse move, and a return that mutated its original would
...               erase the very fact it exists to record.
...
...               The returnable cap of AC-STOCK-022 is absolute in this phase: there is no
...               override, by any caller. See [INV-STK-307] for the note on the supervisor
...               override the AC anticipates but that is deliberately not built.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Ensure Correction Fixtures
Test Tags         inventory    stock_flows    return


*** Test Cases ***
A Done Transfer Can Raise A Return
    [Documentation]    AC-STOCK-021. The return is a new transfer, in draft, pointing back at the
    ...    one it reverses.
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    60
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ${original}=    Set Variable    ${resp.json()}[data]
    Set Suite Variable    ${RETURN_ORIGINAL_ID}    ${transfer_id}
    Set Suite Variable    ${RETURN_ORIGINAL_UPDATED_AT}    ${original.get('updated_at')}

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/create_return
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

The Return Is A New Draft Transfer Pointing At The Original
    [Documentation]    STOCK-INV-011: a return is a new transaction, never a reopening.
    ${return_id}=    Find Return Of Transfer    ${RETURN_ORIGINAL_ID}
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${return_id}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    draft
    ...    msg=A return re-enters the normal lifecycle as a draft (F5, BR §4.2.10.4 step 8)
    Should Be Equal    ${item}[return_of_id]    ${RETURN_ORIGINAL_ID}
    Set Suite Variable    ${RETURN_TRANSFER_ID}    ${return_id}

The Return Reverses The Direction Of Travel
    [Documentation]    A return of a receipt is a delivery: the goods go back where they came
    ...    from. Getting this wrong produces a transfer moving stock the way it already went.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${RETURN_ORIGINAL_ID}
    ${original}=    Set Variable    ${resp.json()}[data]
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${RETURN_TRANSFER_ID}
    ${reverse}=    Set Variable    ${resp.json()}[data]

    Should Be Equal    ${reverse}[source_location_id]    ${original}[destination_location_id]
    ...    msg=The return must draw from where the original delivered
    Should Be Equal    ${reverse}[destination_location_id]    ${original}[source_location_id]
    ...    msg=The return must deliver back to where the original drew from
    Should Be Equal    ${reverse}[operation_code]    outgoing
    ...    msg=A return of an incoming receipt is an outgoing movement

The Original Transfer Is Untouched
    [Documentation]    AC-STOCK-010, STOCK-INV-005 and BR §4.2.10.5. Not edited, not reopened, not
    ...    cancelled. This is the assertion the whole design of the return exists to satisfy.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}/${RETURN_ORIGINAL_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    done
    ...    msg=Raising a return must not reopen or cancel the original
    Should Be Equal    ${item.get('updated_at')}    ${RETURN_ORIGINAL_UPDATED_AT}
    ...    msg=Raising a return must not write to the original at all

The Return Carries A Move Linked To The One It Reverses
    [Documentation]    BR §4.2.10.5: history must read original move → reverse move. The link is
    ...    also what lets the next return tell what has already come back.
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 200} }}
    Response Status Should Be    ${resp}    200
    ${linked}=    Set Variable    ${0}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('transfer_id') == $RETURN_TRANSFER_ID and $item.get('origin_move_id')
            ${linked}=    Evaluate    ${linked} + 1
        END
    END
    Should Be True    ${linked} > 0
    ...    msg=Every move of a return must point at the move it reverses

Returning More Than Was Shipped Is Refused
    [Documentation]    AC-STOCK-022, with no override. The cap is computed from what the original
    ...    actually completed, never from what it demanded (BR §4.2.10.3).
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 200} }}
    ${origin_move}=    Set Variable    ${EMPTY}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('transfer_id') == $RETURN_ORIGINAL_ID
            ${origin_move}=    Set Variable    ${item}[id]
        END
    END
    Should Not Be Empty    ${origin_move}

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${RETURN_ORIGINAL_ID}/create_return
    ...    json=${{ {'lines': [{'move_id': $origin_move, 'quantity': '9999'}]} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A return beyond the returnable quantity must be refused, with no override

Returning A Transfer That Is Not Done Is Refused
    [Documentation]    BR §4.2.10.2. There is nothing to send back until the goods have gone out.
    [Tags]    negative
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${INVENTORY_LOCATION_ID}    5

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/create_return
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Only a completed transfer can be returned


*** Keywords ***
Find Return Of Transfer
    [Documentation]    Locates the reverse transfer raised against an original.
    [Arguments]    ${original_id}
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 200} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('return_of_id') == $original_id    RETURN    ${item}[id]
    END
    Fail    No return transfer was raised against '${original_id}'
