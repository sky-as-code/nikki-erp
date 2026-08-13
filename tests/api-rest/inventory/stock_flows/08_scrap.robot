*** Settings ***
Documentation     The scrap path: raise a draft, execute it, and find the goods gone from usable
...               stock (BR §4.2.9).
...
...               The two rules worth the suite are AC-STOCK-019 — a scrap must generate a
...               movement, not merely mark a document — and AC-STOCK-020: a done scrap cannot be
...               deleted, because the movement it generated is permanent and deleting the
...               document would leave that movement unexplained.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Location Under Test
...               AND    Ensure Product Variant Under Test
...               AND    Ensure Correction Fixtures
Test Tags         inventory    stock_flows    scrap


*** Test Cases ***
A Draft Scrap Changes Nothing
    [Documentation]    BR §4.2.9.3. Raising the document is a statement of intent; only Do Scrap
    ...    moves goods.
    ${transfer_id}    ${move_id}=    Receive Stock Into Location
    ...    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}    50
    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
    ${before}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}

    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'source_location_id': $STOCK_LOCATION_ID, 'scrap_location_id': $SCRAP_LOCATION_ID, 'quantity': '5', 'reason_code': 'damage', 'org_id': $INV_ORG_ID} }}
    ${scrap_id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    Should Be Equal As Numbers    ${after}    ${before}
    ...    msg=A draft scrap must not move stock

    Set Suite Variable    ${SCRAP_ID}    ${scrap_id}
    Set Suite Variable    ${SCRAP_BEFORE}    ${before}

A New Scrap Is Draft And Numbered By The Server
    [Documentation]    The number identifies the document on paperwork, so a client-chosen one
    ...    could collide with another's or impersonate its reference.
    ${resp}=    GET On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    draft
    Should Not Be Empty    ${item}[scrap_number]
    ...    msg=The server must generate a scrap number

Do Scrap Removes The Goods From Usable Stock
    [Documentation]    AC-STOCK-019. The balance falls by exactly the scrapped quantity.
    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}/do_scrap
    ...    json=${{ {} }}    expected_status=any
    Response Status Should Be    ${resp}    200

    ${after}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${expected}=    Evaluate    ${SCRAP_BEFORE} - 5
    Should Be Equal As Numbers    ${after}    ${expected}
    ...    msg=Do Scrap must reduce on-hand by exactly the scrapped quantity

A Done Scrap Records Its Movement
    [Documentation]    The audit trail from document to stock. A done scrap with no move id would
    ...    be a write-off nobody could trace.
    ${resp}=    GET On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[data]
    Should Be Equal    ${item}[status]    done
    Should Not Be Equal    ${item.get('move_id')}    ${None}
    ...    msg=A done scrap must record the move it generated
    Should Not Be Equal    ${item.get('completed_at')}    ${None}
    ...    msg=A done scrap must be stamped with when it ran

A Done Scrap Cannot Be Edited
    [Documentation]    BR §4.2.9.4. Editing a completed scrap would rewrite the description of a
    ...    movement that has already happened.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}
    ${etag}=    Set Variable    ${resp.json()}[data][etag]
    ${resp}=    PUT On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}
    ...    json=${{ {'reason_code': 'data_error', 'etag': $etag} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A completed scrap must not be editable

A Done Scrap Cannot Be Deleted
    [Documentation]    AC-STOCK-020 and BR §4.2.9.6. The remedy for a mistaken scrap is a reverse
    ...    movement, never making its history disappear.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${STOCK_SCRAP_API}/${SCRAP_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A completed scrap must not be deletable

A Draft Scrap Can Be Deleted
    [Documentation]    BR §4.2.9.6. Nothing has happened yet, so there is no side effect to
    ...    orphan.
    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'source_location_id': $STOCK_LOCATION_ID, 'scrap_location_id': $SCRAP_LOCATION_ID, 'quantity': '1', 'org_id': $INV_ORG_ID} }}
    ${draft_id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    DELETE On Session    api    ${STOCK_SCRAP_API}/${draft_id}    expected_status=any
    Response Status Should Be    ${resp}    200

Scrapping More Than Is Available Is Refused
    [Documentation]    The check is against available, not on-hand: reserved stock is promised to
    ...    a transfer, and scrapping it would leave that transfer unable to ship.
    [Tags]    negative
    ${on_hand}=    Read Stock On Hand    ${PRODUCT_VARIANT_ID}    ${STOCK_LOCATION_ID}
    ${too_much}=    Evaluate    ${on_hand} + 1000
    ${qty}=    Convert To String    ${too_much}

    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'source_location_id': $STOCK_LOCATION_ID, 'scrap_location_id': $SCRAP_LOCATION_ID, 'quantity': $qty, 'org_id': $INV_ORG_ID} }}
    ${scrap_id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}/${scrap_id}/do_scrap
    ...    json=${{ {} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Scrapping more than is available must be refused

A Zero Quantity Scrap Is Refused
    [Documentation]    A scrap of nothing is a document asserting a movement that never happened.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_SCRAP_API}
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID, 'source_location_id': $STOCK_LOCATION_ID, 'scrap_location_id': $SCRAP_LOCATION_ID, 'quantity': '0', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A scrap must remove a quantity greater than zero
