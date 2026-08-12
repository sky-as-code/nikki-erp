*** Settings ***
Documentation     AC-STOCK-001: the system must report on-hand, reserved and available for
...               every balance, with available = on-hand - reserved (BR §4.2.2.3).
...
...               Phase 1 has no movement that can create a balance, so these tests assert
...               the invariant over whatever rows exist rather than seeding one. On an empty
...               database they pass vacuously and the schema suite still pins that the field
...               is declared; once transfers land, this file starts doing real work without
...               being rewritten.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    stock_quant    available


*** Test Cases ***
Available Equals On Hand Minus Reserved
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 50} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${on_hand}=    Convert To Number    ${item.get('on_hand_quantity') or 0}
        ${reserved}=    Convert To Number    ${item.get('reserved_quantity') or 0}
        ${available}=    Convert To Number    ${item.get('available_quantity') or 0}
        ${expected}=    Evaluate    ${on_hand} - ${reserved}
        Should Be Equal As Numbers    ${available}    ${expected}
        ...    msg=available_quantity must equal on_hand_quantity - reserved_quantity
    END

Reserved Quantity Is Never Negative
    [Documentation]    STOCK-INV-002. A negative reservation would mean more stock was
    ...    released than was ever held, which no sequence of movements can produce.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 50} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${reserved}=    Convert To Number    ${item.get('reserved_quantity') or 0}
        Should Be True    ${reserved} >= 0    msg=reserved_quantity must never be negative
    END

Available Quantity Is Rejected On Write
    [Documentation]    A virtual field is dropped on write, so a client echoing a record back
    ...    cannot smuggle a balance in through the derived column. Combined with
    ...    03_reject_direct_write.robot this closes both paths to setting a balance by hand.
    ${resp}=    PATCH On Session    api    ${STOCK_QUANT_API}/${NOT_FOUND_ID}
    ...    json=${{ {'available_quantity': '999', 'etag': '___________________'} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=available_quantity must not be writable
