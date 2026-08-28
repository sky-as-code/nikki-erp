*** Settings ***
Documentation     Purchase-module helpers shared by the order, agreement and configuration
...               suites.
...
...               The resources form a dependency chain: an order line hangs off an order, an
...               agreement line off an agreement, and every one of them needs a vendor, a
...               currency, a product and a unit that live in OTHER modules. Each fixture
...               keyword is therefore lazy and pulls in what it depends on, which keeps every
...               suite file runnable standalone.
...
...               Paths are the dynamic resource engine's: /v1/purchase/{schema_name}, with the
...               schema name spelled exactly as the Go constant. The engine asserts permissions
...               using the schema name as the resource code, so a path that drifts from the
...               schema is REFUSED rather than 404'd — which reads as a permission problem and
...               sends the reader looking in the wrong place.
...
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource
Resource          ${CURDIR}/contacts.resource


*** Variables ***
${PURCHASE_ORDER_API}       /v1/purchase/purchase_order
${PURCHASE_ORDER_LINE_API}  /v1/purchase/purchase_order_line
${AGREEMENT_API}            /v1/purchase/purchase_agreement
${AGREEMENT_LINE_API}       /v1/purchase/purchase_agreement_line
${PURCHASE_CONFIG_API}      /v1/purchase/purchase_configuration
${SOURCING_GROUP_API}       /v1/purchase/purchase_sourcing_group
${AUDIT_EVENT_API}          /v1/purchase/purchase_audit_event
${VENDOR_PRICE_API}         /v1/purchase/purchase_vendor_product_price
${PURCHASE_SCHEMA_DIR}      ${CURDIR}/schemas/purchase


*** Keywords ***
Ensure Purchase Org
    [Documentation]    Every purchase resource declares org_id required-for-create. ORG_ID from
    ...    the environment wins when set; otherwise the first organization the API returns is
    ...    used, which keeps the suite app-agnostic instead of hardcoding a seeded ULID.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${PURCHASE_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${PURCHASE_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to own the Purchase fixtures; set ORG_ID in the environment
    Set Global Variable    ${PURCHASE_ORG_ID}    ${items}[0][id]

Ensure Purchase Buyer
    [Documentation]    buyer_id is required on both the order and the agreement. It names a
    ...    user, so the authenticated account is used — which is also the truthful value, since
    ...    that account is the one raising the order.
    ${resolved}=    Get Variable Value    ${PURCHASE_BUYER_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_user    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=No user exists to be the buyer of a purchase order
    Set Global Variable    ${PURCHASE_BUYER_ID}    ${items}[0][id]

Ensure Purchase Vendor
    [Documentation]    D3: "is a vendor" means "has a contacts_vendor_profile row in this
    ...    organization". A bare party is NOT a vendor, so this creates the party AND its
    ...    profile — an order naming a profile-less party is refused, which is the rule the
    ...    vendor suite pins.
    ${resolved}=    Get Variable Value    ${PURCHASE_VENDOR_ID}    ${EMPTY}
    IF    $resolved    RETURN
    Ensure Purchase Org
    ${party_id}    ${etag}=    Create Party    Robot Purchase Vendor    company
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $party_id, 'org_id': $PURCHASE_ORG_ID, 'status': 'active'} }}
    Response Status Should Be    ${resp}    201
    Set Global Variable    ${PURCHASE_VENDOR_ID}    ${party_id}
    Set Global Variable    ${PURCHASE_VENDOR_PROFILE_ID}    ${resp.json()}[id]

Ensure Purchase Currency
    [Documentation]    Resolves an active currency for the order. The ISO-4217 seed of
    ...    [PUR-006] guarantees one exists, so an empty result is a broken environment rather
    ...    than an empty database.
    ...
    ...    The currency drives rounding through decimal_places ([PUR-018]), which is why the
    ...    suites pin a specific one rather than letting the vendor default decide.
    [Arguments]    ${code}=USD
    ${resolved}=    Get Variable Value    ${PURCHASE_CURRENCY_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/essential/essential_currency
    ...    params=${{ {'size': 1, 'graph': '{"if":["code","eq","%s"]}' % $code} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=Currency ${code} is missing; the ISO-4217 seed has not been applied
    Set Global Variable    ${PURCHASE_CURRENCY_ID}    ${items}[0][id]

Ensure Purchase Product
    [Documentation]    Resolves a purchasable product variant and the unit its stock is
    ...    counted in. Both are needed for a product line: PUR-R8 checks the line's unit against
    ...    the product's inventory unit, and the two must share a UoM category.
    ...
    ...    Nothing is created here. Products are Inventory's master data, and a purchase suite
    ...    that created its own would be asserting against a fixture rather than against the
    ...    catalog a real order draws from.
    ${resolved}=    Get Variable Value    ${PURCHASE_VARIANT_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/inventory/inventory_product_variant
    ...    params=${{ {'size': 1, 'fields': 'id,product_template_id,template_purchase_ok'} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=No product variant exists to put on a purchase line
    Set Global Variable    ${PURCHASE_VARIANT_ID}    ${items}[0][id]
    Set Global Variable    ${PURCHASE_TEMPLATE_ID}    ${items}[0][product_template_id]

Ensure Purchase Uom
    [Documentation]    Resolves a unit of measure for the line. Purchase is the first consumer
    ...    of Essential's UoM ([PUR-017]), so an empty result here means the UoM seed is missing
    ...    rather than that the suite is wrong.
    ${resolved}=    Get Variable Value    ${PURCHASE_UOM_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/essential/essential_uom    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=No unit of measure exists to price a purchase line against
    Set Global Variable    ${PURCHASE_UOM_ID}    ${items}[0][id]

Ensure Purchase Fixtures
    [Documentation]    Everything an order needs, in one call. Suites use this rather than
    ...    listing the five separately, so a new dependency is added in one place.
    Ensure Purchase Org
    Ensure Purchase Buyer
    Ensure Purchase Vendor
    Ensure Purchase Currency

# --- Purchase order ----------------------------------------------------------------------

Create Purchase Order
    [Documentation]    Creates an order and returns (id, etag).
    ...
    ...    Note what is NOT sent: code, status, is_locked, vendor_acknowledged,
    ...    approval_required and the three totals are all server-owned. Sending them would be
    ...    harmless — they are overwritten rather than rejected — but leaving them out is what
    ...    the create suite asserts, so the fixture must not model the opposite.
    [Arguments]    ${priority}=normal
    Ensure Purchase Fixtures
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}
    ...    json=${{ {'vendor_id': $PURCHASE_VENDOR_ID, 'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID, 'priority': $priority} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Purchase Order Under Test
    [Documentation]    The order the order suite exercises and that the line suite hangs its
    ...    lines off.
    ${id}=    Get Variable Value    ${PURCHASE_ORDER_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Purchase Order
    Set Global Variable    ${PURCHASE_ORDER_ID}    ${id}
    Set Global Variable    ${PURCHASE_ORDER_ETAG}    ${etag}

Create Purchase Order Line
    [Documentation]    Adds one product line to an order and returns (id, etag).
    ...
    ...    subtotal, total and inventory_quantity are NOT sent: they are computed and no_update.
    ...    tax_amount IS sent, because it is an input rather than a calculation (D9) — there is
    ...    no tax engine and the requirement forbids building one.
    [Arguments]    ${order_id}    ${quantity}=10    ${unit_price}=25.00    ${tax}=0
    ...    ${discount}=0    ${sequence}=1
    Ensure Purchase Product
    Ensure Purchase Uom
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_LINE_API}
    ...    json=${{ {'purchase_order_id': $order_id, 'sequence': int($sequence), 'line_type': 'product', 'product_variant_id': $PURCHASE_VARIANT_ID, 'uom_id': $PURCHASE_UOM_ID, 'quantity': $quantity, 'unit_price': $unit_price, 'discount_percent': $discount, 'tax_amount': $tax, 'org_id': $PURCHASE_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Purchase Order Line Under Test
    [Documentation]    The line the line suite exercises, on the order under test.
    ${id}=    Get Variable Value    ${PURCHASE_ORDER_LINE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Purchase Order Under Test
    ${id}    ${etag}=    Create Purchase Order Line    ${PURCHASE_ORDER_ID}
    Set Global Variable    ${PURCHASE_ORDER_LINE_ID}    ${id}
    Set Global Variable    ${PURCHASE_ORDER_LINE_ETAG}    ${etag}

Create Confirmable Purchase Order
    [Documentation]    An order with one priced line, which is the minimum a confirm accepts:
    ...    an order with no product line is refused because it would commit the business to
    ...    buying nothing.
    [Arguments]    ${quantity}=10    ${unit_price}=25.00
    ${id}    ${etag}=    Create Purchase Order
    Create Purchase Order Line    ${id}    ${quantity}    ${unit_price}
    ${refreshed}=    GET On Session    api    ${PURCHASE_ORDER_API}/${id}
    Response Status Should Be    ${refreshed}    200
    RETURN    ${id}    ${refreshed.json()}[etag]

Confirm Purchase Order
    [Documentation]    Confirms an order and returns the response, without asserting on it:
    ...    callers assert their own expectation, which for the alternatives suite is a REFUSAL
    ...    rather than a success.
    [Arguments]    ${order_id}    ${alternative_choice}=${EMPTY}
    ${body}=    Create Dictionary
    IF    $alternative_choice
        Set To Dictionary    ${body}    alternative_choice    ${alternative_choice}
    END
    ${resp}=    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/confirm
    ...    json=${body}    expected_status=any
    RETURN    ${resp}

# --- Purchase agreement ------------------------------------------------------------------

Create Purchase Agreement
    [Documentation]    Creates an agreement and returns (id, etag). code and status are
    ...    server-owned, exactly as on the order.
    [Arguments]    ${agreement_type}=blanket_order
    Ensure Purchase Fixtures
    ${resp}=    POST On Session    api    ${AGREEMENT_API}
    ...    json=${{ {'agreement_type': $agreement_type, 'vendor_id': $PURCHASE_VENDOR_ID, 'buyer_id': $PURCHASE_BUYER_ID, 'currency_id': $PURCHASE_CURRENCY_ID, 'org_id': $PURCHASE_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Agreement Under Test
    [Documentation]    The agreement the agreement suite exercises.
    ${id}=    Get Variable Value    ${AGREEMENT_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Purchase Agreement
    Set Global Variable    ${AGREEMENT_ID}    ${id}
    Set Global Variable    ${AGREEMENT_ETAG}    ${etag}

Create Agreement Line
    [Documentation]    Adds one committed line to an agreement and returns (id, etag).
    [Arguments]    ${agreement_id}    ${quantity}=100    ${unit_price}=20.00    ${sequence}=1
    Ensure Purchase Product
    Ensure Purchase Uom
    ${resp}=    POST On Session    api    ${AGREEMENT_LINE_API}
    ...    json=${{ {'purchase_agreement_id': $agreement_id, 'sequence': int($sequence), 'product_variant_id': $PURCHASE_VARIANT_ID, 'uom_id': $PURCHASE_UOM_ID, 'quantity': $quantity, 'unit_price': $unit_price, 'org_id': $PURCHASE_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Agreement Line Under Test
    [Documentation]    The line the agreement line suite exercises.
    ${id}=    Get Variable Value    ${AGREEMENT_LINE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Agreement Under Test
    ${id}    ${etag}=    Create Agreement Line    ${AGREEMENT_ID}
    Set Global Variable    ${AGREEMENT_LINE_ID}    ${id}
    Set Global Variable    ${AGREEMENT_LINE_ETAG}    ${etag}

Create Confirmed Agreement
    [Documentation]    An agreement with one line, confirmed — the state create_rfq requires.
    ...    A draft cannot raise one, because its prices were never agreed with anybody.
    ${id}    ${etag}=    Create Purchase Agreement
    Create Agreement Line    ${id}
    ${resp}=    POST On Session    api    ${AGREEMENT_API}/${id}/confirm
    Response Status Should Be    ${resp}    200
    RETURN    ${id}

# --- Audit -------------------------------------------------------------------------------

Get Audit Events For
    [Documentation]    Reads the audit trail of one record. PUR-R6 requires exactly one event
    ...    per state-changing action, so the suites count as well as inspect.
    [Arguments]    ${entity_id}
    ${graph}=    Set Variable    {"if":["entity_id","eq","${entity_id}"]}
    ${resp}=    GET On Session    api    ${AUDIT_EVENT_API}
    ...    params=${{ {'size': 100, 'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items]

Audit Trail Should Record
    [Documentation]    Asserts the trail carries exactly one event for the given action, and
    ...    returns it. "Exactly one" is the point: a second event for one confirm would mean
    ...    the operation ran twice, which is the failure PUR-R6's one-transaction rule prevents.
    [Arguments]    ${entity_id}    ${action}
    ${events}=    Get Audit Events For    ${entity_id}
    ${matching}=    Create List
    FOR    ${event}    IN    @{events}
        IF    $event['action'] == $action
            Append To List    ${matching}    ${event}
        END
    END
    Length Should Be    ${matching}    1
    ...    msg=Expected exactly one '${action}' audit event for ${entity_id}, found ${matching}
    RETURN    ${matching}[0]

# --- Cleanup -----------------------------------------------------------------------------

Delete Purchase Seed Data
    [Documentation]    Suite teardown. Only fixtures created on the fly are removed, and only
    ...    the ones that CAN be: an order is deletable solely from cancelled (BR 24), so an
    ...    order left confirmed is cancelled first rather than force-deleted.
    ...
    ...    Failures are swallowed. A teardown that fails masks the real failure of the test that
    ...    preceded it, and a stranded fixture is a smaller problem than a hidden one.
    Run Keyword And Ignore Error    Delete Purchase Order Fixture    ${PURCHASE_ORDER_ID}
    Run Keyword And Ignore Error    Delete Agreement Fixture    ${AGREEMENT_ID}

Delete Purchase Order Fixture
    [Documentation]    Cancels then deletes one order, skipping quietly when it was never
    ...    created.
    [Arguments]    ${order_id}
    IF    not $order_id    RETURN
    POST On Session    api    ${PURCHASE_ORDER_API}/${order_id}/cancel
    ...    json=${{ {'reason': 'robot teardown'} }}    expected_status=any
    DELETE On Session    api    ${PURCHASE_ORDER_API}/${order_id}    expected_status=any

Delete Agreement Fixture
    [Documentation]    Cancels then deletes one agreement. A draft is deletable outright, but
    ...    cancelling first covers the confirmed case too.
    [Arguments]    ${agreement_id}
    IF    not $agreement_id    RETURN
    POST On Session    api    ${AGREEMENT_API}/${agreement_id}/cancel
    ...    json=${{ {'reason': 'robot teardown'} }}    expected_status=any
    DELETE On Session    api    ${AGREEMENT_API}/${agreement_id}    expected_status=any

# --- Business violations ------------------------------------------------------------------
# The key/message pairs are server contracts emitted by modules/purchase/domain/services.
# Keep them verbatim; the frontend localizes by key through purchase.json. Asserting the KEY
# rather than a bare 4xx is what makes these tests worth having: a refusal for the wrong
# reason passes a status-code check and fails this one.

Response Should Be Purchase Violation
    [Documentation]    Asserts the response is a business refusal carrying exactly the given
    ...    violation key. The message is not asserted — it is prose for a human reader and may
    ...    be reworded, where the key is what a client branches on.
    [Arguments]    ${resp}    ${key}
    Response Status Should Be    ${resp}    400
    ${keys}=    Create List
    FOR    ${item}    IN    @{resp.json()}
        Append To List    ${keys}    ${item}[key]
    END
    Should Contain    ${keys}    ${key}
    ...    msg=Expected a '${key}' violation; the server sent ${keys}

Response Should Be Client Error
    [Documentation]    A weaker check for the cases where the refusal is the schema's rather
    ...    than the module's — a missing required field, say — and the exact key belongs to the
    ...    common error contract instead of to Purchase.
    [Arguments]    ${resp}
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=Expected a client error, got ${resp.status_code}
