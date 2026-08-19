*** Settings ***
Documentation     Payment & Invoice helpers shared by the order and invoice suites.
...               The fixtures form a short chain — an order needs a payment method, which
...               needs a currency; an invoice line needs an invoice — so each keyword is
...               lazy and pulls in what it depends on, keeping every suite file runnable
...               standalone.
...
...               Two things about this module shape the suites. Orders are not created
...               through the engine's plain create: `create_payment` is what mints one,
...               and it talks to a gateway. Every gateway defaults to ENABLED: false, so
...               against a stock deployment that action is expected to fail as a *client*
...               error rather than a 500 — which is itself the contract worth pinning, and
...               is all of it that belongs in CI. Real gateway traffic is not CI material.
...
...               All keywords expect an authorized session with alias "api".
Library           Collections
Library           String
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${PAYMENT_METHOD_API}       /v1/paymentinvoice/paymentinvoice_payment_method
${ORDER_API}                /v1/paymentinvoice/paymentinvoice_order
${TRANSACTION_API}          /v1/paymentinvoice/paymentinvoice_transaction
${INVOICE_API}              /v1/paymentinvoice/paymentinvoice_invoice
${INVOICE_LINE_API}         /v1/paymentinvoice/paymentinvoice_invoice_line
${CURRENCY_API}             /v1/essential/essential_currency
${PAYINV_SCHEMA_DIR}        ${CURDIR}/schemas/paymentinvoice


*** Keywords ***
Ensure Payment Invoice Org
    [Documentation]    Every model in this module declares org_id required-for-create.
    ...    ORG_ID from the environment wins when set; otherwise the first organization the
    ...    API returns is used, so the suite is not tied to a seeded ULID.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${PAYINV_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${PAYINV_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to own the Payment fixtures; set ORG_ID in the environment
    Set Global Variable    ${PAYINV_ORG_ID}    ${items}[0][id]

Ensure Payment Currency
    [Documentation]    Money in this module is denominated by a currency that lives in
    ...    Essential, referenced by a plain ulid with no foreign key. An existing currency is
    ...    reused when there is one — creating a second VND would be the odd thing to do — and
    ...    one is minted only when the deployment has none.
    ${resolved}=    Get Variable Value    ${PAYINV_CURRENCY_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    ${CURRENCY_API}    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    len($items) > 0
        Set Global Variable    ${PAYINV_CURRENCY_ID}    ${items}[0][id]
        RETURN
    END
    #    The code is ISO 4217 shaped and the schema enforces ^[A-Z]{3}$, so it is generated
    #    from upper-case letters only — the shared Unique Suffix mixes in digits, which the
    #    regex rejects.
    ${code}=    Generate Random String    3    [UPPER]
    ${resp}=    POST On Session    api    ${CURRENCY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': 'Robot Currency'}, 'symbol': 'R', 'decimal_places': 2, 'is_active': True} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PAYINV_CURRENCY_ID}    ${id}
    Set Global Variable    ${PAYINV_CREATED_CURRENCY_ID}    ${id}

Ensure Payment Method Under Test
    [Documentation]    The payment method an order names. adapter_code points at the gateway
    ...    adapter that serves it, which is how offering a gateway under a second merchant
    ...    account is a row rather than a release. The code used here names a gateway that is
    ...    real but disabled by default, which is exactly the state create_payment must refuse
    ...    cleanly.
    ${resolved}=    Get Variable Value    ${PAYMENT_METHOD_ID}    ${EMPTY}
    IF    $resolved    RETURN
    Ensure Payment Currency
    ${code}=    Unique Payment Code    method
    #    A payment method is deliberately not org-scoped: the merchant contract behind it
    #    belongs to the deployment, not to one organization within it.
    ${resp}=    POST On Session    api    ${PAYMENT_METHOD_API}
    ...    json=${{ {'code': $code, 'adapter_code': 'momo', 'name': {'en-US': 'Robot Momo'}, 'currency_id': $PAYINV_CURRENCY_ID, 'is_active': True, 'min_amount': '1000', 'max_amount': '50000000'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PAYMENT_METHOD_ID}    ${id}
    Set Global Variable    ${PAYMENT_METHOD_ETAG}    ${etag}
    Set Global Variable    ${PAYMENT_METHOD_CODE}    ${code}

Ensure Invoice Under Test
    [Documentation]    The draft invoice the issue suite closes. It is created through the
    ...    plain engine create, unlike an order: an invoice is authored, not collected.
    ${resolved}=    Get Variable Value    ${INVOICE_ID}    ${EMPTY}
    IF    $resolved    RETURN
    Ensure Payment Invoice Org
    Ensure Payment Currency
    ${partner}=    Unique Display Name    Robot Partner
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    json=${{ {'partner_name': $partner, 'currency_id': $PAYINV_CURRENCY_ID, 'org_id': $PAYINV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INVOICE_ID}    ${id}
    Set Global Variable    ${INVOICE_ETAG}    ${etag}

Create Draft Invoice
    [Documentation]    Mints a further draft and returns its id, for the tests that need an
    ...    invoice of their own rather than the shared one under test.
    Ensure Payment Invoice Org
    Ensure Payment Currency
    ${partner}=    Unique Display Name    Robot Partner
    ${resp}=    POST On Session    api    ${INVOICE_API}
    ...    json=${{ {'partner_name': $partner, 'currency_id': $PAYINV_CURRENCY_ID, 'org_id': $PAYINV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Add Invoice Line
    [Documentation]    Appends one line to an invoice and returns its id. amount is
    ...    deliberately not sent: the issue action recomputes it from quantity x unit_price,
    ...    and several tests turn on that being true.
    [Arguments]    ${invoice_id}    ${quantity}    ${unit_price}    ${tax_rate_percent}=0
    Ensure Payment Invoice Org
    ${description}=    Unique Display Name    Robot Line
    ${resp}=    POST On Session    api    ${INVOICE_LINE_API}
    ...    json=${{ {'invoice_id': $invoice_id, 'description': $description, 'quantity': $quantity, 'unit_price': $unit_price, 'tax_rate_percent': $tax_rate_percent, 'amount': '0', 'org_id': $PAYINV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Unique Payment Code
    [Documentation]    `code` is unique per org on the payment method, so it carries the same
    ...    {words}+{hash} treatment as every other unique column.
    [Arguments]    ${words}=pay
    ${suffix}=    Unique Suffix
    RETURN    ${words}_${suffix}

Delete Payment Invoice Fixture
    [Documentation]    Best-effort delete of one record, clearing the global that held it so a
    ...    later suite cannot reuse a dead id.
    [Arguments]    ${api}    ${variable_name}
    ${id}=    Get Variable Value    ${${variable_name}}    ${EMPTY}
    IF    $id
        DELETE On Session    api    ${api}/${id}    expected_status=any
        Set Global Variable    ${${variable_name}}    ${EMPTY}
    END

Delete Payment Invoice Seed Data
    [Documentation]    Best-effort removal of the fixtures these suites created on the fly,
    ...    referencing-before-referenced. Lines point at their invoice and transactions at
    ...    their order, so those unwind first; the payment method is pointed at by both orders
    ...    and transactions, so it goes last. The currency is removed only when this run
    ...    created it — a deployment's own currency is not ours to delete.
    Delete Payment Invoice Fixture    ${INVOICE_LINE_API}    INVOICE_LINE_ID
    Delete Payment Invoice Fixture    ${INVOICE_API}    INVOICE_ID
    Delete Payment Invoice Fixture    ${ORDER_API}    ORDER_ID
    Delete Payment Invoice Fixture    ${PAYMENT_METHOD_API}    PAYMENT_METHOD_ID
    Delete Payment Invoice Fixture    ${CURRENCY_API}    PAYINV_CREATED_CURRENCY_ID
