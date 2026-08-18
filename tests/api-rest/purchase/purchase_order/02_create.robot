*** Settings ***
Documentation     Creating a purchase order. AC-01 and AC-02.
...
...               What this suite is really pinning is the boundary between what a client
...               chooses and what the server owns. A client picks the vendor, the buyer, the
...               currency and the priority; the server mints the code, forces the status to
...               rfq, zeroes the totals and starts every flag false. Getting that boundary
...               wrong is how an order gets created already confirmed, or already approved.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    purchase_order    create


*** Test Cases ***
Create Purchase Order Succeeds
    [Documentation]    AC-01: a buyer can raise a request for quotation against a vendor.
    Ensure Purchase Order Under Test
    Should Not Be Empty    ${PURCHASE_ORDER_ID}

New Order Starts As A Request For Quotation
    [Documentation]    AC-01. Every order starts as an RFQ and earns the rest. One created
    ...    `purchase_order` would be a commitment with no confirmation behind it.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[status]    rfq

Server Mints The Order Code
    [Documentation]    The code identifies the order to the vendor, so letting a client pick it
    ...    would let two orders collide or one impersonate another's reference.
    ...
    ...    The prefix is PO for every order regardless of status (PUR-R1): a code that encoded
    ...    the status would have to change when the status did, while the vendor is holding a
    ...    document quoting the old one.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Response Status Should Be    ${resp}    200
    Should Start With    ${resp.json()}[code]    PO-

A Client Supplied Code And Status Are Overwritten Not Rejected
    [Documentation]    They are not part of the request's meaning, and a client echoing a
    ...    record back should not fail for carrying them. What must NOT happen is the server
    ...    honouring them.
    Ensure Purchase Fixtures
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}
    ...    json=${{ {'vendor_id': $PURCHASE_VENDOR_ID, 'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID, 'priority': 'normal', 'code': 'PO-CLIENT-CHOSEN', 'status': 'purchase_order'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${created}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Status Should Be    ${created}    200
    Should Not Be Equal    ${created.json()}[code]    PO-CLIENT-CHOSEN
    Should Be Equal    ${created.json()}[status]    rfq
    [Teardown]    Delete Purchase Order Fixture    ${id}

New Order Totals Start At Zero
    [Documentation]    D8: the totals are a summary of lines that do not exist yet. They are
    ...    stamped rather than omitted because they are required_for_create.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal As Numbers    ${body}[untaxed_amount]    0
    Should Be Equal As Numbers    ${body}[tax_amount]        0
    Should Be Equal As Numbers    ${body}[total_amount]      0

New Order Flags Start False
    [Documentation]    is_locked and vendor_acknowledged are things that happen TO an order
    ...    later; approval_required is decided at confirm time against the org's configuration
    ...    and the order's total, neither of which is knowable at create.
    Ensure Purchase Order Under Test
    ${resp}=    GET On Session    api    ${PURCHASE_ORDER_API}/${PURCHASE_ORDER_ID}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Not Be True    ${body}[is_locked]
    Should Not Be True    ${body}[vendor_acknowledged]
    Should Not Be True    ${body}[approval_required]

Create Requires A Vendor
    [Documentation]    vendor_id is required_for_create: an order with no vendor is a
    ...    commitment to buy from nobody.
    Ensure Purchase Fixtures
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}
    ...    json=${{ {'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID, 'priority': 'normal'} }} expected_status=any
    Response Should Be Client Error    ${resp}

Create Refuses A Party That Is Not A Vendor
    [Documentation]    AC-02, D3. "Is a vendor" means "has a contacts_vendor_profile row in
    ...    this organization", which is a checkable fact. A bare party is a contact, not a
    ...    supplier, and ordering from one would commit the business to a counterparty nobody
    ...    ever qualified.
    Ensure Purchase Fixtures
    ${party_id}    ${etag}=    Create Party    Robot Not A Vendor    company
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}
    ...    json=${{ {'vendor_id': $party_id, 'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID, 'priority': 'normal'} }} expected_status=any
    Response Should Be Client Error    ${resp}
