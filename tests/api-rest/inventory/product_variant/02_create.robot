*** Settings ***
Documentation     Creating Product Variants, including the combination uniqueness of
...               BR-PROD-VAR-002. The first test saves the variant under test
...               (${PRODUCT_VARIANT_ID}/${PRODUCT_VARIANT_ETAG}) consumed by the later
...               suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_variant    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${key}=    Unique Code    comb
    ${sku}=    Unique Code    sku
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $key, 'sku': $sku, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_VARIANT_ID}    ${id}
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_VARIANT_COMBINATION}    ${key}

Create With Empty Combination Succeeds
    [Documentation]    BR §4.5 / AC-PROD-008: a template with no variant-generating
    ...    attributes still has exactly one variant, whose combination key is the empty
    ...    string. Empty is a real and normal combination, not a missing value — treating it
    ...    as absent would leave such a template with no transactable SKU at all.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Empty Combination Template
    ${sku}=    Unique Code    emptysku
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $template_id, 'combination_key': '', 'sku': $sku, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[combination_key]    ${EMPTY}
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${id}    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}    expected_status=any

Create With Duplicate Combination Fails
    [Documentation]    BR-PROD-VAR-002 / AC-PROD-012: a template holds at most one variant
    ...    per attribute combination, or a transaction line could not say which SKU it meant.
    ...    The engine turns the composite unique into a field-level business error rather
    ...    than letting the constraint surface as a 500.
    [Tags]    negative
    ${sku}=    Unique Code    dupsku
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $PRODUCT_VARIANT_COMBINATION, 'sku': $sku, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Combination Error    ${resp}

Create Same Combination On Another Template Succeeds
    [Documentation]    The uniqueness of BR-PROD-VAR-002 is scoped to the template, not
    ...    global: "Black / Large" under two different product lines are two different SKUs.
    ...    A global constraint would make the second product line unbuildable.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Sibling Template
    ${sku}=    Unique Code    siblingsku
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $template_id, 'combination_key': $PRODUCT_VARIANT_COMBINATION, 'sku': $sku, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${id}    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}    expected_status=any

Create Defaults To Active And Materialized
    [Documentation]    BR §6.2.2: a variant created directly is a real SKU, so it defaults to
    ...    active and materialized. is_materialized is false only for a dynamic-mode
    ...    combination that is known valid but has not been used yet.
    ${key}=    Unique Code    defcomb
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $key, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[status]    active
    Should Be Equal    ${item}[is_materialized]    ${True}
    Should Be True    ${{ not $item.get('archive_source') }}
    ...    msg=A live variant must carry no archive_source stamp
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${id}    expected_status=any

Create Inherits The Template Fields
    [Documentation]    BR §7.6 / AC-PROD-032: the template's name and classification reach a
    ...    variant read through the virtual template_* fields, without being stored on it.
    ...    A consumer must never have to join the two halves itself.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Not Be Empty    ${item}[template_name]
    ...    msg=A variant must inherit its template's name rather than storing one
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${item}[etag]

Create With Missing Required Fields Fails
    [Documentation]    status is required_for_create too, but it declares a default, and the
    ...    schema applies a default before it reports a field as missing. required_for_create
    ...    is what makes the column NOT NULL; it does not oblige the caller to send a value
    ...    the schema can supply. Its defaulting is covered by Create Defaults To Active And
    ...    Materialized.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    product_template_id    combination_key    org_id

Create With Not Found Template Fails
    [Documentation]    A variant cannot exist without the template it inherits from — there
    ...    would be nothing to resolve its name, category or capability flags against.
    [Tags]    negative
    ${key}=    Unique Code    orphan
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $NOT_FOUND_ID, 'combination_key': $key, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A variant must not be creatable against a template that does not exist

Create With Invalid Status Fails
    [Tags]    negative
    ${key}=    Unique Code    badstatus
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $key, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A variant has no draft state; status is active or discontinued only

Create With Inherited Field Fails
    [Documentation]    AC-PROD-004: `name` belongs to the template. Accepting it here would
    ...    create the second source of truth the split exists to prevent.
    [Tags]    negative
    ${key}=    Unique Code    namedcomb
    ${name}=    Unique Display Name    Robot Named Variant
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $key, 'status': 'active', 'org_id': $INV_ORG_ID, 'name': {'en-US': $name}} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    name

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    data={ "sku": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${key}=    Unique Code    nonexist
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'combination_key': $key, 'status': 'active', 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
