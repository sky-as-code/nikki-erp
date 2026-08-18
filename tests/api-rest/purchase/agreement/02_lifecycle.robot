*** Settings ***
Documentation     The agreement lifecycle: create, confirm, create_rfq, close, cancel, archive.
...               AC-15 through AC-20, §32 to §46.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_agreement    lifecycle


*** Test Cases ***
Create Agreement Starts As A Draft
    [Documentation]    AC-15. Server-owned code and status, exactly as on the order: an
    ...    agreement created `confirmed` would be a commitment nobody made.
    Ensure Agreement Under Test
    ${resp}=    GET On Session    api    ${AGREEMENT_API}/${AGREEMENT_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[status]    draft
    Should Start With    ${resp.json()}[code]    PA-

Confirm Makes The Agreement Live
    [Documentation]    AC-16, §38. Only a confirmed agreement can have orders drawn against it.
    ${id}    ${etag}=    Create Purchase Agreement
    Create Agreement Line    ${id}
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${id}/confirm
    Response Status Should Be    ${resp}    200
    ${agreement}=    GET On Session    api    ${AGREEMENT_API}/${id}
    Should Be Equal    ${agreement.json()}[status]    confirmed
    Audit Trail Should Record    ${id}    confirm
    [Teardown]    Delete Agreement Fixture    ${id}

Confirm Is Refused On An Agreement With No Line
    [Documentation]    AC-16. A blanket order with no line commits to no quantity at no price,
    ...    which is not an agreement any vendor could act on.
    ${id}    ${etag}=    Create Purchase Agreement
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${id}/confirm    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_agreement.no_lines
    [Teardown]    Delete Agreement Fixture    ${id}

Create Rfq Raises An Order From The Agreement
    [Documentation]    AC-17, §40. The agreement's lines become the order's, at the agreed
    ...    quantities and prices — that is the point of a blanket order. What the order does NOT
    ...    inherit is the agreement's status: it starts as an ordinary draft RFQ that happens to
    ...    have been pre-filled, and must be confirmed like any other.
    ${agreement_id}=    Create Confirmed Agreement
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${agreement_id}/create_rfq
    Response Status Should Be    ${resp}    200
    ${order_id}=    Set Variable    ${resp.json()}[id]
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Response Status Should Be    ${order}    200
    Should Be Equal    ${order.json()}[status]        rfq
    Should Be Equal    ${order.json()}[agreement_id]  ${agreement_id}
    Should Be Equal As Numbers    ${order.json()}[total_amount]    2000.00
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${order_id}
    ...    AND    Delete Agreement Fixture    ${agreement_id}

Create Rfq Is Refused On A Draft Agreement
    [Documentation]    §40. A draft's prices were never agreed with anybody, so an order quoting
    ...    them would quote terms nobody committed to.
    ${id}    ${etag}=    Create Purchase Agreement
    Create Agreement Line    ${id}
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${id}/create_rfq    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_agreement.not_confirmed
    [Teardown]    Delete Agreement Fixture    ${id}

Close Is Refused While Orders Are Open Against The Agreement
    [Documentation]    AC-18 and the non-obvious part of §42. Closing would strand the open
    ...    orders: they reference terms no longer in force, and nothing would say whether those
    ...    terms still apply to goods on their way.
    ${agreement_id}=    Create Confirmed Agreement
    ${created}=    POST On Session    api    ${AGREEMENT_API}/${agreement_id}/create_rfq
    ${order_id}=    Set Variable    ${created.json()}[id]
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${agreement_id}/close    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_agreement.has_open_orders
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${order_id}
    ...    AND    Delete Agreement Fixture    ${agreement_id}

Close Succeeds Once The Open Orders Are Settled
    [Documentation]    AC-18. Orders already confirmed against it are untouched: closing says
    ...    "no more orders from here", not "the ones already placed are void".
    ${agreement_id}=    Create Confirmed Agreement
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${agreement_id}/close
    Response Status Should Be    ${resp}    200
    ${agreement}=    GET On Session    api    ${AGREEMENT_API}/${agreement_id}
    Should Be Equal    ${agreement.json()}[status]    closed
    Audit Trail Should Record    ${agreement_id}    close
    [Teardown]    Delete Agreement Fixture    ${agreement_id}

Cancel Calls The Agreement Off
    [Documentation]    AC-19, §43.
    ${id}    ${etag}=    Create Purchase Agreement
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${id}/cancel    json=${{ {'reason': 'terms not agreed'} }}
    Response Status Should Be    ${resp}    200
    ${agreement}=    GET On Session    api    ${AGREEMENT_API}/${id}
    Should Be Equal    ${agreement.json()}[status]    cancelled
    ${event}=    Audit Trail Should Record    ${id}    cancel
    Should Be Equal    ${event}[reason]    terms not agreed
    [Teardown]    Delete Agreement Fixture    ${id}

Archive And Restore Use The Built In Set Archived
    [Documentation]    AC-20, §44 and §45. They are NOT separate actions: restore is the same
    ...    power applied in reverse, so splitting them would let a role archive agreements it
    ...    could not bring back. This is a deliberate correction to the plan's §4 table.
    ${id}    ${etag}=    Create Purchase Agreement
    ${resp}=    PUT On Session    api    ${AGREEMENT_API}/${id}/archived    json=${{ {'etag': $etag, 'is_archived': ${True}} }}
    Response Status Should Be    ${resp}    200
    ${archived}=    GET On Session    api    ${AGREEMENT_API}/${id}
    Should Be True    ${archived.json()}[is_archived]
    ${restore}=    PUT On Session    api    ${AGREEMENT_API}/${id}/archived    json=${{ {'etag': $archived.json()['etag'], 'is_archived': ${False}} }}
    Response Status Should Be    ${restore}    200
    [Teardown]    Delete Agreement Fixture    ${id}

A Draft Agreement Is Deletable Where A Draft Order Is Not
    [Documentation]    BR 46. An agreement's code is not quoted to a vendor until it is
    ...    confirmed, which is the whole reason the two rules differ.
    ${id}    ${etag}=    Create Purchase Agreement
    ${resp}=    DELETE On Session    api    ${AGREEMENT_API}/${id}
    Response Should Be Delete Success    ${resp}

Delete Is Refused On A Confirmed Agreement
    ${agreement_id}=    Create Confirmed Agreement
    ${resp}=    DELETE On Session    api    ${AGREEMENT_API}/${agreement_id}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_agreement.not_deletable
    [Teardown]    Delete Agreement Fixture    ${agreement_id}
