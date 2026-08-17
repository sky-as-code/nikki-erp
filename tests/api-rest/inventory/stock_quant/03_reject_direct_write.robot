*** Settings ***
Documentation     AC-STOCK-002 and BR §3.3/§4.2.2.6: no business API may change a stock
...               balance directly. Every change to on-hand must trace back to a completed
...               movement, so create, update and delete are all refused on this resource —
...               corrections go through an inventory adjustment, transfer or scrap.
...
...               This is the suite that would catch the single most damaging regression in
...               the module: a balance that can be set by hand is a balance no report can
...               explain and no audit can reconstruct.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Location Under Test
...               AND    Ensure Product Variant Under Test
Test Tags         inventory    stock_quant    negative


*** Test Cases ***
Create Is Refused
    [Documentation]    A well-formed create is used deliberately: the request must be
    ...    refused because balances are not client-writable, not because it was malformed.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'location_id': $INVENTORY_LOCATION_ID, 'lot_ref': '', 'package_ref': '', 'owner_ref': '', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A stock balance must not be creatable through the API

Create With Quantities Is Refused
    [Documentation]    The form that matters most: "set product X at location Y to 100".
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'location_id': $INVENTORY_LOCATION_ID, 'lot_ref': '', 'package_ref': '', 'owner_ref': '', 'on_hand_quantity': '100', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=Setting an on-hand quantity directly must not be accepted

Update Is Refused
    [Documentation]    Uses a not-found id on purpose: whichever answer comes back, a 200 is
    ...    the one outcome that must never happen, since it would mean the action is live.
    ${resp}=    PATCH On Session    api    ${STOCK_QUANT_API}/${NOT_FOUND_ID}
    ...    json=${{ {'on_hand_quantity': '50', 'etag': '___________________'} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A stock balance must not be updatable through the API

Delete Is Refused
    ${resp}=    DELETE On Session    api    ${STOCK_QUANT_API}/${NOT_FOUND_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A stock balance must not be deletable through the API

Archive Is Not Offered
    [Documentation]    BR §4.2.2.2: a quant carries no is_archived, so the archive action
    ...    must not exist for it — archive is not a substitute for a corrective movement.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A stock balance must not expose an archive action
