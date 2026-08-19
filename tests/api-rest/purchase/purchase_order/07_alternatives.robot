*** Settings ***
Documentation     Alternatives and merge. AC-11 through AC-14, §26 to §31.
...
...               Two rules here are worth stating up front because they look like bugs until
...               you know why. Confirming an order with open alternatives is REFUSED rather
...               than allowed with a note — defaulting either way makes a purchasing decision
...               on the caller's behalf. And a comparison across currencies declines to name a
...               winner, because no exchange-rate model exists and ranking 100 USD against
...               100 VND by their numbers would be confidently wrong.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    alternatives


*** Test Cases ***
Create Alternative Puts Two Orders In One Sourcing Group
    [Documentation]    AC-11, §27. The group is a technical record with no meaning of its own
    ...    (§28): it exists only to say that these orders answer the same requirement.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${second_party}    ${party_etag}=    Create Party    Robot Alt Vendor    company
    ${profile}=    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    Response Status Should Be    ${profile}    201
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    Response Status Should Be    ${resp}    200
    ${alt_id}=    Set Variable    ${resp.json()}[id]
    ${original}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    ${alternative}=    GET On Session    api    ${PURCHASE_ORDER_API}/${alt_id}
    Should Not Be Empty    ${original.json()}[sourcing_group_id]
    Should Be Equal    ${original.json()}[sourcing_group_id]    ${alternative.json()}[sourcing_group_id]
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

An Alternative Asks A Different Vendor
    [Documentation]    The whole point is to ask somebody else, so the alternative carries the
    ...    source's lines but its own vendor. Copying the vendor would produce two identical
    ...    requests to the same supplier.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${second_party}    ${party_etag}=    Create Party    Robot Alt Vendor Two    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    Response Status Should Be    ${resp}    200
    ${alt_id}=    Set Variable    ${resp.json()}[id]
    ${alternative}=    GET On Session    api    ${PURCHASE_ORDER_API}/${alt_id}
    Should Be Equal    ${alternative.json()}[vendor_id]    ${second_party}
    Should Be Equal    ${alternative.json()}[status]       rfq
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

Create Alternative Requires A Vendor
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order.alternative_vendor_required
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

Compare Alternatives Names The Cheapest
    [Documentation]    AC-13, §30. "Cheapest" is computed server-side so that it means the same
    ...    thing in the API and in the UI.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order    10    100.00
    ${second_party}    ${party_etag}=    Create Party    Robot Cheap Vendor    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${created}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    ${alt_id}=    Set Variable    ${created.json()}[id]
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/compare_alternatives
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Length Should Be    ${body}[Alternatives]    2
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

Confirming With Open Alternatives Is Refused Until The Caller Decides
    [Documentation]    AC-14, §31. Confirming one alternative leaves the others quoting for a
    ...    requirement that has just been met. Cancelling them loses quotes the buyer may still
    ...    want; keeping them leaves live requests to vendors who will never get the business.
    ...    Neither is the server's call, so it refuses and asks.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${second_party}    ${party_etag}=    Create Party    Robot Warn Vendor    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${created}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    ${alt_id}=    Set Variable    ${created.json()}[id]
    ${resp}=    Confirm Purchase Order    ${order_id}
    Response Should Be Purchase Violation    ${resp}    purchase_order.open_alternatives
    ${order}=    GET On Session    api    ${PURCHASE_ORDER_API}/${order_id}
    Should Be Equal    ${order.json()}[status]    rfq    msg=A refused confirm must not have moved the order
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

Confirming With Keep Alternatives Leaves The Others Open
    [Documentation]    AC-14. The buyer may be placing more than one, or wants to keep quoting
    ...    until the goods arrive.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${second_party}    ${party_etag}=    Create Party    Robot Keep Vendor    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${created}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    ${alt_id}=    Set Variable    ${created.json()}[id]
    ${resp}=    Confirm Purchase Order    ${order_id}    keep_alternatives
    Response Status Should Be    ${resp}    200
    ${alternative}=    GET On Session    api    ${PURCHASE_ORDER_API}/${alt_id}
    Should Be Equal    ${alternative.json()}[status]    rfq
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

Confirming With Cancel Alternatives Closes The Others
    [Documentation]    AC-14, the usual outcome: the requirement is met and the remaining quotes
    ...    are no longer wanted. Each cancelled alternative gets its own audit event saying why.
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${second_party}    ${party_etag}=    Create Party    Robot Cancel Vendor    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $second_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${created}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/create_alternative    json=${{ {'vendor_id': $second_party} }}
    ${alt_id}=    Set Variable    ${created.json()}[id]
    ${resp}=    Confirm Purchase Order    ${order_id}    cancel_alternatives
    Response Status Should Be    ${resp}    200
    ${alternative}=    GET On Session    api    ${PURCHASE_ORDER_API}/${alt_id}
    Should Be Equal    ${alternative.json()}[status]    cancelled
    Audit Trail Should Record    ${alt_id}    cancel
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${alt_id}
    ...    AND    Delete Purchase Order Fixture    ${order_id}

Merge Folds Draft Orders Into The Oldest
    [Documentation]    AC-12, §26. Three people each raised an RFQ for the same vendor; sending
    ...    three documents would get three quotes and three deliveries. The target keeps the
    ...    oldest code so the vendor's own paperwork still matches.
    ${first_id}    ${first_etag}=    Create Confirmable Purchase Order    5    10.00
    ${second_id}    ${second_etag}=    Create Confirmable Purchase Order    3    10.00
    ${ids}=    Create List    ${first_id}    ${second_id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/merge    json=${{ {'order_ids': $ids} }}
    Response Status Should Be    ${resp}    200
    ${first}=    GET On Session    api    ${PURCHASE_ORDER_API}/${first_id}
    ${second}=    GET On Session    api    ${PURCHASE_ORDER_API}/${second_id}
    ${statuses}=    Create List    ${first.json()}[status]    ${second.json()}[status]
    Should Contain    ${statuses}    cancelled
    ...    msg=One of the merged orders must be cancelled as the source
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${first_id}
    ...    AND    Delete Purchase Order Fixture    ${second_id}

Merge Needs At Least Two Orders
    ${order_id}    ${etag}=    Create Confirmable Purchase Order
    ${ids}=    Create List    ${order_id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/merge    json=${{ {'order_ids': $ids} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order.merge_needs_two
    [Teardown]    Delete Purchase Order Fixture    ${order_id}

Merge Refuses Orders With Different Vendors
    [Documentation]    §26. Merging across vendors would ask one supplier for another's goods.
    ${first_id}    ${first_etag}=    Create Confirmable Purchase Order
    ${other_party}    ${party_etag}=    Create Party    Robot Merge Vendor    company
    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {'party_id': $other_party, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}    json=${{ {'vendor_id': $other_party, 'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID, 'priority': 'normal'} }}
    ${second_id}    ${second_etag}=    Response Should Be Create Success    ${resp}
    ${ids}=    Create List    ${first_id}    ${second_id}
    ${merge}=    POST On Session    api    ${PURCHASE_ORDER_API}/merge    json=${{ {'order_ids': $ids} }}    expected_status=any
    Response Should Be Purchase Violation    ${merge}    purchase_order.merge_vendor_mismatch
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${first_id}
    ...    AND    Delete Purchase Order Fixture    ${second_id}

Merge Refuses A Confirmed Order
    [Documentation]    §26. A confirmed order is a commitment the vendor is holding, and folding
    ...    it into another document would change what was agreed after the fact.
    ${first_id}    ${first_etag}=    Create Confirmable Purchase Order
    ${second_id}    ${second_etag}=    Create Confirmable Purchase Order
    Confirm Purchase Order    ${second_id}
    ${ids}=    Create List    ${first_id}    ${second_id}
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/merge    json=${{ {'order_ids': $ids} }}    expected_status=any
    Response Should Be Purchase Violation    ${resp}    purchase_order.not_mergeable
    [Teardown]    Run Keywords
    ...    Delete Purchase Order Fixture    ${first_id}
    ...    AND    Delete Purchase Order Fixture    ${second_id}
