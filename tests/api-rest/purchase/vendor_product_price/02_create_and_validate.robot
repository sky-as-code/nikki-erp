*** Settings ***
Documentation     Creating a vendor price, and the write-time checks that guard it (section 23).
...
...               Every reference this resource holds is a plain ulid with no foreign key behind
...               it — the vendor belongs to Contacts, the product to Inventory, the unit to
...               Essential. Nothing in the database stops a row naming something that does not
...               exist, is blocked, or belongs to a different product. That is the price of not
...               declaring foreign keys across a module boundary, and these checks are what pays
...               it: they run at write time, where a useful message can still be produced.
Resource          resources/purchase.resource
Suite Setup       Set Up Vendor Price Fixtures
Suite Teardown    Delete Vendor Price Fixtures
Test Tags         purchase    vendor_product_price


*** Variables ***
${VENDOR_PRICE_SCHEMA}      ${PURCHASE_SCHEMA_DIR}/purchase_vendor_product_price.json
@{CREATED_PRICE_IDS}


*** Test Cases ***
Create A Template Wide Vendor Price
    [Documentation]    No product_variant_id: the quote covers every variant of the template. This
    ...    is the ordinary case, not a fallback — most suppliers quote a product, not a colour.
    ${id}    ${etag}=    Create Vendor Price    min_quantity=1    unit_price=250000
    Should Not Be Empty    ${id}

    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/${id}
    Response Should Match Schema    ${resp}    ${VENDOR_PRICE_SCHEMA}    200
    Should Be Equal    ${resp.json()}[vendor_id]    ${PURCHASE_VENDOR_ID}
    Should Be Equal    ${resp.json()}[product_template_id]    ${PURCHASE_TEMPLATE_ID}

Create A Variant Specific Vendor Price
    [Documentation]    PRICE-INV-018: naming a variant makes the quote more specific, and it beats
    ...    the template-wide one at resolution.
    ${id}    ${etag}=    Create Vendor Price    min_quantity=1    unit_price=235000
    ...    product_variant_id=${PURCHASE_VARIANT_ID}

    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/${id}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[product_variant_id]    ${PURCHASE_VARIANT_ID}

A New Vendor Price Is Not Archived
    [Documentation]    A quote is live the moment it is recorded. Defaulting to archived would make
    ...    every new price invisible to resolution until somebody unarchived it.
    ${id}    ${etag}=    Create Vendor Price    min_quantity=1    unit_price=100000

    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/${id}
    Response Status Should Be    ${resp}    200
    Should Not Be True    ${resp.json()}[is_archived]

Quantity Breaks Are Recorded Separately
    [Documentation]    TS-PRICE-06. Three rows, not one row with three prices: each break is its
    ...    own quote with its own validity, and flattening them would make it impossible to expire
    ...    one without the others.
    ${one}    ${etag}=    Create Vendor Price    min_quantity=1      unit_price=250000
    ${ten}    ${etag}=    Create Vendor Price    min_quantity=10     unit_price=240000
    ${hundred}    ${etag}=    Create Vendor Price    min_quantity=100    unit_price=220000

    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}
    ...    params=${{ {'size': 50, 'graph': '{"if":["product_template_id","eq","%s"]}' % $PURCHASE_TEMPLATE_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be True    ${resp.json()}[total] >= 3

A Price For An Unknown Vendor Is Refused
    [Documentation]    The vendor is a ulid with nothing enforcing it. A row naming a party that
    ...    does not exist would look perfectly ordinary until an order tried to resolve through it.
    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}
    ...    json=${{ dict($VENDOR_PRICE_BODY, vendor_id='01ZZZZZZZZZZZZZZZZZZZZZZZZ') }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A price for a non-existent vendor was accepted; nothing else would have caught it

A Price For An Unknown Product Is Refused
    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}
    ...    json=${{ dict($VENDOR_PRICE_BODY, product_template_id='01ZZZZZZZZZZZZZZZZZZZZZZZZ') }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Variant From Another Template Is Refused
    [Documentation]    The two columns must describe ONE product. If they disagree, resolution
    ...    would return a price for something nobody asked about — and the row reads as valid.
    ${other}=    Find A Variant Of Another Template
    Skip If    not $other    msg=Only one product template exists; nothing to mismatch against
    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}
    ...    json=${{ dict($VENDOR_PRICE_BODY, product_variant_id=$other) }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Negative Unit Price Is Refused
    [Documentation]    A vendor does not pay the buyer to take goods away. The bound is on the
    ...    field, so this is really asserting that the bound survived.
    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}
    ...    json=${{ dict($VENDOR_PRICE_BODY, unit_price='-1') }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Archiving A Price Keeps It Readable
    [Documentation]    PRICE-INV-024. An archived quote stops pricing anything new, but an order
    ...    that already resolved through it still names it — so it must remain fetchable by id.
    ...    Readable and usable are different things, and conflating them would strand history.
    ${id}    ${etag}=    Create Vendor Price    min_quantity=1    unit_price=999000

    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}/${id}/set_archived
    ...    json=${{ {'is_archived': True, 'etag': $etag} }}    expected_status=any
    Should Be True    ${resp.status_code} < 400    msg=Archiving a vendor price must be permitted

    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/${id}
    Response Status Should Be    ${resp}    200
    Should Be True    ${resp.json()}[is_archived]


*** Keywords ***
Set Up Vendor Price Fixtures
    [Documentation]    Resolves everything a quote references. Nothing here is created except the
    ...    vendor: products, units and currencies are other modules' master data, and a suite that
    ...    invented its own would assert against a fixture rather than the catalog.
    Create Authorized API Session
    Ensure Purchase Fixtures
    Ensure Purchase Product
    Ensure Purchase Uom
    Set Suite Variable    ${VENDOR_PRICE_BODY}    ${{ {
    ...    'org_id': $PURCHASE_ORG_ID,
    ...    'vendor_id': $PURCHASE_VENDOR_ID,
    ...    'product_template_id': $PURCHASE_TEMPLATE_ID,
    ...    'purchase_uom_id': $PURCHASE_UOM_ID,
    ...    'currency_id': $PURCHASE_CURRENCY_ID,
    ...    'min_quantity': '1',
    ...    'unit_price': '100000'
    ...    } }}

Create Vendor Price
    [Documentation]    Creates one quote and remembers it for teardown. Returns (id, etag).
    [Arguments]    ${min_quantity}=1    ${unit_price}=100000    ${product_variant_id}=${EMPTY}
    ${body}=    Evaluate    dict($VENDOR_PRICE_BODY, min_quantity=str($min_quantity), unit_price=str($unit_price))
    IF    $product_variant_id
        ${body}=    Evaluate    dict($body, product_variant_id=$product_variant_id)
    END
    ${resp}=    POST On Session    api    ${VENDOR_PRICE_API}    json=${body}
    Response Status Should Be    ${resp}    201
    Append To List    ${CREATED_PRICE_IDS}    ${resp.json()}[id]
    RETURN    ${resp.json()}[id]    ${resp.json()}[etag]

Find A Variant Of Another Template
    [Documentation]    Returns a variant belonging to some OTHER template, or empty when the
    ...    environment has only one — in which case the mismatch cannot be constructed and the
    ...    test skips rather than passing vacuously.
    ${resp}=    GET On Session    api    /v1/inventory/inventory_product_variant
    ...    params=${{ {'size': 50, 'fields': 'id,product_template_id'} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item['product_template_id'] != $PURCHASE_TEMPLATE_ID
            RETURN    ${item}[id]
        END
    END
    RETURN    ${EMPTY}

Delete Vendor Price Fixtures
    [Documentation]    Removes every quote this suite created. A vendor price has no dependents —
    ...    an order line records the price it resolved rather than a reference that would dangle —
    ...    so these delete outright rather than needing to be cancelled first.
    ...
    ...    Failures are swallowed: a stranded fixture is a smaller problem than a teardown failure
    ...    masking the real one.
    FOR    ${id}    IN    @{CREATED_PRICE_IDS}
        Run Keyword And Ignore Error
        ...    DELETE On Session    api    ${VENDOR_PRICE_API}/${id}    expected_status=any
    END
