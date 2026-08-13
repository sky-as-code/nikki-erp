*** Settings ***
Documentation     Physical inventory: enter a count, reset it, apply it as an adjustment.
...
...               The rule this suite exists for is the stale-snapshot refusal of BR §4.2.7.4 and
...               AC-STOCK-015: a count is a statement about the balance as it stood when the
...               counter looked, so applying it to a balance that has since moved would write a
...               number nobody ever counted.
...
...               Every case starts with a receipt, because the quant is read-only to clients
...               (AC-STOCK-002) and validating an incoming transfer is the only way stock enters.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Ensure Correction Fixtures
Test Tags         inventory    stock_flows    inventory_count


*** Test Cases ***
Entering A Count Does Not Change On Hand
    [Documentation]    AC-STOCK-014. Recording what a counter found is not a correction; only
    ...    applying it is. If entering a count moved stock, a miscount would be unrecoverable.
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}    40
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/enter_count
    ...    json=${{ {'counted_quantity': '37', 'count_reason_code': 'missing'} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${before}
    ...    msg=Entering a count must not move stock

    Set Suite Variable    ${COUNT_QUANT_ID}    ${quant_id}
    Set Suite Variable    ${COUNT_ON_HAND}    ${before}

Entering A Count Records The Pending Flag And Snapshot
    [Documentation]    The snapshot is taken at entry, not at apply: it is the whole basis of the
    ...    staleness check, and one taken at apply would always match.
    ${flag}=    Read Stock Quant Field    ${COUNT_QUANT_ID}    count_quantity_set
    Should Be True    ${flag}    msg=count_quantity_set must be set once a count is entered
    ${snapshot}=    Read Stock Quant Field    ${COUNT_QUANT_ID}    count_snapshot_quantity
    Should Be Equal As Numbers    ${snapshot}    ${COUNT_ON_HAND}
    ...    msg=The snapshot must record on-hand as it stood when the count was entered

Applying A Count Moves The Balance To What Was Counted
    [Documentation]    BR §4.2.7.4. The variance of -3 is applied by generating a movement, never
    ...    by writing the quant directly (decision F3, INV-STK-R14).
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${COUNT_QUANT_ID}/apply_adjustment
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    37
    ...    msg=Applying a count of 37 must leave the balance at exactly 37

Applying A Count Clears The Pending Count
    [Documentation]    BR §4.2.7.5. The count is resolved, so the worklist must stop asking.
    ${flag}=    Read Stock Quant Field    ${COUNT_QUANT_ID}    count_quantity_set
    Should Not Be True    ${flag}    msg=Applying a count must clear count_quantity_set
    ${last}=    Read Stock Quant Field    ${COUNT_QUANT_ID}    last_count_date
    Should Not Be Equal    ${last}    ${None}    msg=last_count_date must be stamped by apply

Applying A Count Generates An Adjustment Movement
    [Documentation]    The balance changed, so there must be a movement explaining it: an
    ...    adjustment that wrote the quant directly would be a change no report could account for.
    ${resp}=    GET On Session    api    ${STOCK_MOVE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${found}=    Set Variable    ${False}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('is_inventory_adjustment') and $item.get('product_variant_id') == $PRODUCT_VARIANT_ID
            ${found}=    Set Variable    ${True}
        END
    END
    Should Be True    ${found}
    ...    msg=Applying a count must generate a move marked is_inventory_adjustment

Applying With No Pending Count Is Refused
    [Documentation]    The flag is the authority, and it was just cleared by the apply above.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${COUNT_QUANT_ID}/apply_adjustment
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Applying with nothing pending must be refused

A Stale Count Is Refused
    [Documentation]    AC-STOCK-015 and STOCK-INV-008, with the worked example from BR §4.2.7.4:
    ...    a count is entered, stock moves before it is applied, and the apply must refuse rather
    ...    than compute the variance against a balance nobody counted.
    [Tags]    negative
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/enter_count
    ...    json=${{ {'counted_quantity': '30'} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    # The delivery that lands between the count and the apply.
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}    10
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/apply_adjustment
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A count whose snapshot has gone stale must be refused, not applied

Resetting A Count Clears It And Leaves The Balance Alone
    [Documentation]    BR §4.2.7.6. Reset abandons a count entered in error; it is not itself a
    ...    correction, so the balance must be untouched.
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/reset_count
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${flag}=    Read Stock Quant Field    ${quant_id}    count_quantity_set
    Should Not Be True    ${flag}    msg=Reset must clear the pending count
    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${before}
    ...    msg=Reset must not change the balance

A Zero Variance Still Resolves The Count
    [Documentation]    BR §4.2.7.5. A counter who confirms the balance was right has done their
    ...    job; skipping the stamp would leave that balance permanently overdue on the worklist.
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${counted}=    Convert To String    ${on_hand}

    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/enter_count
    ...    json=${{ {'counted_quantity': $counted} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/apply_adjustment
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${flag}=    Read Stock Quant Field    ${quant_id}    count_quantity_set
    Should Not Be True    ${flag}    msg=A zero-variance apply must still clear the pending count
    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${on_hand}
    ...    msg=A zero variance must leave the balance exactly where it was

A Negative Counted Quantity Is Refused
    [Documentation]    Nothing physical is present in a negative amount. Zero is allowed, and is
    ...    tested separately, because "the shelf is empty" is a legitimate count.
    [Tags]    negative
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/enter_count
    ...    json=${{ {'counted_quantity': '-1'} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A negative counted quantity must be refused

Scheduling And Assigning A Count Are Plain Field Writes
    [Documentation]    BR §4.2.8. Cycle counting needs no ledger of its own: the worklist is a
    ...    filtered search over next_count_date, so scheduling is a field write and nothing more.
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/schedule_count
    ...    json=${{ {'next_count_date': '2026-12-01'} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    ${next}=    Read Stock Quant Field    ${quant_id}    next_count_date
    Should Contain    ${next}    2026-12-01

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/${quant_id}/assign_counter
    ...    json=${{ {'count_assigned_user_id': '01HQ9WBZ3XKAAAAAAAAAAAAAAA'} }}    expected_status=any
    Response Status Should Be    ${resp}    200

The Counts Due Worklist Is An Ordinary Filtered Search
    [Documentation]    AC-STOCK-016. The scheduled balance above must be findable by filtering on
    ...    next_count_date — which is the whole cycle-count worklist, with no new entity.
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${due}=    Set Variable    ${0}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('next_count_date')
            ${due}=    Evaluate    ${due} + 1
        END
    END
    Should Be True    ${due} > 0
    ...    msg=At least one balance must carry a next_count_date after being scheduled

Direct Writes To A Balance Are Still Refused
    [Documentation]    AC-STOCK-002 still holds. The count actions write count metadata and never
    ...    on_hand_quantity, so adding them must not have reopened the resource to client writes.
    [Tags]    negative
    ${quant_id}=    Read Stock Quant Id    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${resp}=    PUT On Session    api    ${STOCK_QUANT_API}/${quant_id}
    ...    json=${{ {'on_hand_quantity': '999'} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Client writes to a stock balance must still be refused
