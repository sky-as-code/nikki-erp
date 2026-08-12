*** Settings ***
Documentation     Creating Product Templates. The first test saves the template under test
...               (${PRODUCT_TEMPLATE_ID}/${PRODUCT_TEMPLATE_ETAG}) consumed by the later
...               suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Product Type Under Test
...               AND    Ensure Product Category Under Test
Test Tags         inventory    product_template    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_TEMPLATE_ID}    ${id}
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Create Defaults To Draft Status
    [Documentation]    BR §6.1.2: a new product line starts as draft, not active. A template
    ...    that defaulted to active would be sellable before anyone reviewed it.
    ${name}=    Unique Display Name    Robot Default Status Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[status]    draft
    Should Be Equal    ${item}[is_archived]    ${False}
    ...    msg=A new template must not be born archived
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${id}    expected_status=any

Create Defaults The Capability Flags On
    [Documentation]    BR §6.1.2: sale_ok and purchase_ok default true, so a template is
    ...    transactable unless deliberately restricted. Variants inherit both (BR §7.6), so
    ...    a wrong default here silently propagates to every SKU.
    ${name}=    Unique Display Name    Robot Capability Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[sale_ok]    ${True}
    Should Be Equal    ${item}[purchase_ok]    ${True}
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${id}    expected_status=any

Create With Brand And Defaults Succeeds
    [Documentation]    The optional half of the model: brand plus the fallback dimensions a
    ...    variant inherits when it carries none of its own. Decimals travel as strings —
    ...    a JSON number would be parsed as a float64 and lose precision in transit.
    Ensure Brand Under Test
    ${name}=    Unique Display Name    Robot Full Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'brand_id': $BRAND_ID, 'default_weight': '1.5', 'default_length': '10', 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[brand_id]    ${BRAND_ID}
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${id}    expected_status=any

Create Without Brand Succeeds
    [Documentation]    brand_id is nullable: an unbranded or own-label product is normal,
    ...    so the catalog must not force a brand to exist before a template can.
    ${name}=    Unique Display Name    Robot Unbranded Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be True    ${{ not $item.get('brand_id') }}
    ...    msg=A template created without a brand should have none
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${id}    expected_status=any

Create With Missing Required Fields Fails
    [Documentation]    status is required_for_create too, but it declares a default, and the
    ...    schema applies a default before it reports a field as missing. required_for_create
    ...    is what makes the column NOT NULL; it does not oblige the caller to send a value
    ...    the schema can supply. Its defaulting is covered by Create Defaults To Draft Status.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    name    product_type_id    category_id    org_id

Create With Invalid Status Fails
    [Documentation]    BR §6.1.2: the lifecycle is a closed set. An unknown status would
    ...    leave every consumer guessing whether the product is sellable.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Status Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'bla_bla_status', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=status must reject a value outside draft/active/discontinued

Create With Not Found Product Type Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Orphan Type Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $NOT_FOUND_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A template must not be creatable against a product type that does not exist

Create With Not Found Category Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Orphan Category Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $NOT_FOUND_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A template must not be creatable against a category that does not exist

Create With Variant Owned Field Fails
    [Documentation]    AC-PROD-004: `sku` belongs to the variant. Accepting it here would
    ...    let a caller believe the template carries one.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Sku Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID, 'sku': 'SKU-1'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    sku

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
