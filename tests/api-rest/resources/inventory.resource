*** Settings ***
Documentation     Inventory-module helpers shared by the product suites. The seven core
...               resources form a dependency chain — a template needs a product type and
...               a category, a variant needs a template, an attribute value needs an
...               attribute — so each fixture keyword is lazy and pulls in the ones it
...               depends on. That keeps every suite file runnable standalone.
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${PRODUCT_TYPE_API}         /v1/inventory/inventory_product_type
${PRODUCT_CATEGORY_API}     /v1/inventory/inventory_product_category
${BRAND_API}                /v1/inventory/inventory_brand
${PRODUCT_ATTRIBUTE_API}    /v1/inventory/inventory_product_attribute
${ATTRIBUTE_VALUE_API}      /v1/inventory/inventory_product_attribute_value
${PRODUCT_TEMPLATE_API}     /v1/inventory/inventory_product_template
${PRODUCT_VARIANT_API}      /v1/inventory/inventory_product_variant
${TEMPLATE_ATTRIBUTE_API}           /v1/inventory/inventory_product_template_attribute
${TEMPLATE_ATTRIBUTE_VALUE_API}     /v1/inventory/inventory_product_template_attribute_value
${INVENTORY_LOCATION_API}               /v1/inventory/inventory_location
${WAREHOUSE_API}                    /v1/inventory/inventory_warehouse
${STORAGE_CATEGORY_API}             /v1/inventory/inventory_storage_category
${SUPPLY_RELATION_API}              /v1/inventory/inventory_warehouse_supply_relation
${PUTAWAY_RULE_API}                 /v1/inventory/inventory_putaway_rule
${STOCK_OPERATION_TYPE_API}         /v1/inventory/inventory_stock_operation_type
${STOCK_QUANT_API}                  /v1/inventory/inventory_stock_quant
${STOCK_TRANSFER_API}               /v1/inventory/inventory_stock_transfer
${STOCK_MOVE_API}                   /v1/inventory/inventory_stock_move
${STOCK_MOVE_LINE_API}              /v1/inventory/inventory_stock_move_line
${STOCK_MOVE_DEPENDENCY_API}        /v1/inventory/inventory_stock_move_dependency
${STOCK_SCRAP_API}                  /v1/inventory/inventory_stock_scrap
${STOCK_PRODUCT_CONFIG_API}         /v1/inventory/inventory_stock_product_config
${INVENTORY_SCHEMA_DIR}     ${CURDIR}/schemas/inventory


*** Keywords ***
Ensure Inventory Org
    [Documentation]    Every Inventory model except the two template-attribute junctions
    ...    declares org_id required-for-create. ORG_ID from the environment wins when set;
    ...    otherwise the first organization the API returns is used, which keeps the suite
    ...    app-agnostic instead of hardcoding a seeded ULID.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${INV_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${INV_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to own the Inventory fixtures; set ORG_ID in the environment
    Set Global Variable    ${INV_ORG_ID}    ${items}[0][id]

Unique Code
    [Documentation]    `code` is unique (per org on most of these models), so it carries the
    ...    same {words}+{hash} treatment as every other unique column.
    [Arguments]    ${words}=inv
    ${suffix}=    Unique Suffix
    RETURN    ${words}_${suffix}

# --- Master data ----------------------------------------------------------------------

Ensure Product Type Under Test
    [Documentation]    The product type the product_type suite exercises, and that every
    ...    template hangs off — product_type_id is required-for-create on a template.
    ${id}=    Get Variable Value    ${PRODUCT_TYPE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Goods Type
    ${code}=    Unique Code    goods
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'supports_stock': True} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_TYPE_ID}    ${id}
    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_TYPE_CODE}    ${code}

Ensure Product Category Under Test
    [Documentation]    The category the product_category suite exercises and every template
    ...    references. Created as a root category, so 03_update can hang a child off it and
    ...    exercise the cycle rule against a real two-level tree.
    ${id}=    Get Variable Value    ${PRODUCT_CATEGORY_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Root Category
    ${code}=    Unique Code    cat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_CATEGORY_ID}    ${id}
    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_CATEGORY_CODE}    ${code}

Ensure Child Product Category
    [Documentation]    A child of the category under test. The cycle rule needs a real
    ...    parent-child pair: pointing the parent at its own descendant is the case
    ...    assertNoCategoryCycle exists to reject.
    ${id}=    Get Variable Value    ${CHILD_CATEGORY_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Product Category Under Test
    ${name}=    Unique Display Name    Robot Child Category
    ${code}=    Unique Code    childcat
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${CHILD_CATEGORY_ID}    ${id}
    Set Global Variable    ${CHILD_CATEGORY_ETAG}    ${etag}

Ensure Brand Under Test
    [Documentation]    The brand the brand suite exercises. Optional on a template
    ...    (brand_id is nullable), so nothing else depends on it.
    ${id}=    Get Variable Value    ${BRAND_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Brand
    ${code}=    Unique Code    brand
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${BRAND_ID}    ${id}
    Set Global Variable    ${BRAND_ETAG}    ${etag}
    Set Global Variable    ${BRAND_CODE}    ${code}

# --- Attributes -----------------------------------------------------------------------

Ensure Product Attribute Under Test
    [Documentation]    An `option`/`instant` attribute — the only combination that takes
    ...    part in variant generation (BR §6.5.3), so the generation suite can use it.
    ${id}=    Get Variable Value    ${PRODUCT_ATTRIBUTE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Color Attribute
    ${code}=    Unique Code    color
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ID}    ${id}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_CODE}    ${code}

Create Attribute Value
    [Documentation]    Creates one value of the given attribute and returns its id. Used
    ...    both by the attribute_value suite and by the 2x2 generation fixture.
    [Arguments]    ${attribute_id}    ${words}=Robot Value
    Ensure Inventory Org
    ${name}=    Unique Display Name    ${words}
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attribute_id, 'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Attribute Value Under Test
    [Documentation]    The value the attribute_value suite exercises.
    ${id}=    Get Variable Value    ${ATTRIBUTE_VALUE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Product Attribute Under Test
    ${id}    ${etag}=    Create Attribute Value    ${PRODUCT_ATTRIBUTE_ID}    Robot Black
    Set Global Variable    ${ATTRIBUTE_VALUE_ID}    ${id}
    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${etag}

# --- Template and variant -------------------------------------------------------------

Create Product Template
    [Documentation]    Creates a template against the shared type and category and returns
    ...    (id, etag). Callers that need their own template — the archive-cascade and
    ...    delete-guard suites each consume one destructively — use this instead of the
    ...    shared "under test" one.
    [Arguments]    ${words}=Robot Template
    Ensure Product Type Under Test
    Ensure Product Category Under Test
    ${name}=    Unique Display Name    ${words}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Product Template Under Test
    [Documentation]    The template the product_template suite exercises and the variant
    ...    suite hangs its variants off.
    ${id}=    Get Variable Value    ${PRODUCT_TEMPLATE_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Product Template    Robot Template
    Set Global Variable    ${PRODUCT_TEMPLATE_ID}    ${id}
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Create Product Variant
    [Documentation]    Creates a variant of the given template with an explicit combination
    ...    key, and returns (id, etag). The key is passed rather than generated because the
    ...    uniqueness rule (BR-PROD-VAR-002) is keyed on it and the tests must control it.
    [Arguments]    ${template_id}    ${combination_key}=${EMPTY}    ${words}=sku
    Ensure Inventory Org
    ${sku}=    Unique Code    ${words}
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}
    ...    json=${{ {'product_template_id': $template_id, 'combination_key': $combination_key, 'sku': $sku, 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Product Variant Under Test
    [Documentation]    The variant the product_variant suite exercises. Its combination key
    ...    is unique per execution so repeated runs against the same database do not trip
    ...    the (product_template_id, combination_key) composite unique.
    ${id}=    Get Variable Value    ${PRODUCT_VARIANT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Product Template Under Test
    ${key}=    Unique Code    comb
    ${id}    ${etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}
    Set Global Variable    ${PRODUCT_VARIANT_ID}    ${id}
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_VARIANT_COMBINATION}    ${key}

Attach Attribute To Template
    [Documentation]    Scopes an attribute to a template (BR §6.7) and returns the
    ...    junction id — the id the allowed-value rows point at, not the attribute's.
    [Arguments]    ${template_id}    ${attribute_id}
    ${resp}=    POST On Session    api    ${TEMPLATE_ATTRIBUTE_API}
    ...    json=${{ {'product_template_id': $template_id, 'attribute_id': $attribute_id} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Allow Attribute Value On Template
    [Documentation]    Adds one attribute value to a template attribute's allowed subset.
    [Arguments]    ${template_attribute_id}    ${attribute_value_id}
    ${resp}=    POST On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    json=${{ {'template_attribute_id': $template_attribute_id, 'attribute_value_id': $attribute_value_id} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Ensure Two By Two Template
    [Documentation]    A template configured with two INSTANT attributes of two values each
    ...    — the 2x2 fixture BR §8.2 / AC-PROD-011 describes, whose generation must yield
    ...    exactly four variants. Built fresh rather than shared, because generation counts
    ...    are only meaningful against a template nothing else has touched.
    ${id}=    Get Variable Value    ${MATRIX_TEMPLATE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${id}    ${etag}=    Create Product Template    Robot Matrix Template
    Set Global Variable    ${MATRIX_TEMPLATE_ID}    ${id}
    Set Global Variable    ${MATRIX_TEMPLATE_ETAG}    ${etag}

    FOR    ${slot}    IN    A    B
        ${name}=    Unique Display Name    Robot Matrix Attribute ${slot}
        ${code}=    Unique Code    matrix${slot}
        ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID} }}
        ${attr_id}    ${attr_etag}=    Response Should Be Create Success    ${resp}
        Set Global Variable    ${MATRIX_ATTRIBUTE_${slot}_ID}    ${attr_id}
        ${tpl_attr_id}=    Attach Attribute To Template    ${id}    ${attr_id}
        FOR    ${index}    IN RANGE    2
            ${value_id}    ${value_etag}=    Create Attribute Value    ${attr_id}    Robot Matrix ${slot}${index}
            Allow Attribute Value On Template    ${tpl_attr_id}    ${value_id}
        END
    END

# --- Bulk seeding ---------------------------------------------------------------------

Ensure Seeded Product Templates
    [Documentation]    Bulk-creates sample templates for Search/Exists once per execution.
    ...    Every 10th carries "Lead" in the name so content-dependent search filters match
    ...    even on an otherwise empty database. Left in the DB by design.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_TEMPLATE_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Product Type Under Test
    Ensure Product Category Under Test
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Template ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Template ${index}
        END
        ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}
        ...    json=${{ {'name': {'en-US': $name}, 'product_type_id': $PRODUCT_TYPE_ID, 'category_id': $PRODUCT_CATEGORY_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_TEMPLATE_IDS}    ${ids}

Ensure Seeded Product Categories
    [Documentation]    Bulk-creates sample categories for Search/Exists, all roots so no
    ...    seed can participate in a cycle the tree rule would then have to reject.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_CATEGORY_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Inventory Org
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Category ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Category ${index}
        END
        ${code}=    Unique Code    seedcat
        ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_CATEGORY_IDS}    ${ids}

Ensure Seeded Product Variants
    [Documentation]    Bulk-creates sample variants of the template under test. Each gets a
    ...    distinct combination key, since (product_template_id, combination_key) is unique.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_VARIANT_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Product Template Under Test
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${key}=    Unique Code    seedcomb${index}
        ${id}    ${etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}    seedsku
        Append To List    ${ids}    ${id}
    END
    Set Global Variable    ${SEEDED_VARIANT_IDS}    ${ids}

Ensure Seeded Brands
    [Documentation]    Bulk-creates sample brands for Search/Exists.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_BRAND_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Inventory Org
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Brand ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Brand ${index}
        END
        ${code}=    Unique Code    seedbrand
        ${resp}=    POST On Session    api    ${BRAND_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_BRAND_IDS}    ${ids}

# --- Teardown -------------------------------------------------------------------------

Delete Inventory Fixture
    [Documentation]    Best-effort delete of one record, clearing the global that held it
    ...    so a later suite cannot reuse a dead id.
    [Arguments]    ${api}    ${variable_name}
    ${id}=    Get Variable Value    ${${variable_name}}    ${EMPTY}
    IF    $id
        DELETE On Session    api    ${api}/${id}    expected_status=any
        Set Global Variable    ${${variable_name}}    ${EMPTY}
    END

Archive Inventory Fixture
    [Documentation]    Best-effort archive of one record, for the resources a delete cannot
    ...    reach. A warehouse owns the locations it created for itself, and those refuse to be
    ...    archived while the warehouse is live, so the warehouse is withdrawn instead of
    ...    removed — which is what archiving is for. Clears the global either way.
    [Arguments]    ${api}    ${variable_name}
    ${id}=    Get Variable Value    ${${variable_name}}    ${EMPTY}
    IF    $id
        ${resp}=    GET On Session    api    ${api}/${id}    expected_status=any
        IF    ${resp.status_code} == 200
            ${etag}=    Set Variable    ${resp.json()}[etag]
            POST On Session    api    ${api}/${id}/archived
            ...    json=${{ {'is_archived': True, 'etag': $etag} }}    expected_status=any
        END
        Set Global Variable    ${${variable_name}}    ${EMPTY}
    END

Delete Matrix Template Fixture
    [Documentation]    Removes the 2x2 generation fixture. Its variants were created by the
    ...    generator rather than by a test, so they are found by search instead of by a
    ...    remembered id, and they must go before the template that owns them.
    ${id}=    Get Variable Value    ${MATRIX_TEMPLATE_ID}    ${EMPTY}
    IF    not $id    RETURN
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100, 'graph': '{"if":["product_template_id", "=", "' + $id + '"]}'} }}
    ...    expected_status=any
    IF    ${resp.status_code} == 200
        FOR    ${item}    IN    @{resp.json()}[items]
            DELETE On Session    api    ${PRODUCT_VARIANT_API}/${item}[id]    expected_status=any
        END
    END
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${id}    expected_status=any
    Set Global Variable    ${MATRIX_TEMPLATE_ID}    ${EMPTY}

# --- Stock ----------------------------------------------------------------------------

Ensure Inventory Location Under Test
    [Documentation]    The internal location the stock suites exercise. Internal is the only
    ...    type that holds company-owned stock, so it is what a quant hangs off.
    ${id}=    Get Variable Value    ${INVENTORY_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Inventory Location
    ${code}=    Unique Code    stkloc
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INVENTORY_LOCATION_ID}    ${id}
    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${etag}
    Set Global Variable    ${INVENTORY_LOCATION_CODE}    ${code}

Ensure Stock Operation Type Under Test
    [Documentation]    The receipt operation type the stock_operation_type suite exercises.
    ${id}=    Get Variable Value    ${STOCK_OPERATION_TYPE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Location Under Test
    ${name}=    Unique Display Name    Robot Receipt Type
    ${code}=    Unique Code    stkop
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'at_confirmation', 'backorder_policy': 'ask', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ID}    ${id}
    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}
    Set Global Variable    ${STOCK_OPERATION_TYPE_CODE}    ${code}

Ensure Stock Destination Location Under Test
    [Documentation]    A second internal location, so a transfer has somewhere to move stock to.
    ...    Source and destination must differ (BR §4.2.3.4), so the movement suites cannot reuse
    ...    the single location the Phase 1 suites share.
    ${id}=    Get Variable Value    ${STOCK_DEST_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Stock Destination
    ${code}=    Unique Code    stkdst
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_DEST_LOCATION_ID}    ${id}

Ensure Stock Transfer Under Test
    [Documentation]    An internal draft transfer between the two locations under test, with one
    ...    move for the variant under test. Internal rather than incoming, because both ends must
    ...    be real locations for the balance assertions to have something to read on each side.
    ${id}=    Get Variable Value    ${STOCK_TRANSFER_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Location Under Test
    Ensure Stock Destination Location Under Test
    Ensure Internal Operation Type Under Test
    Ensure Product Variant Under Test
    ${id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    Set Global Variable    ${STOCK_TRANSFER_ID}    ${id}
    Set Global Variable    ${STOCK_TRANSFER_ETAG}    ${etag}
    ${move_id}=    Add Stock Move    ${id}    ${PRODUCT_VARIANT_ID}    10
    Set Global Variable    ${STOCK_MOVE_ID}    ${move_id}

Ensure Internal Operation Type Under Test
    [Documentation]    An internal-direction operation type reserving at confirmation, which is
    ...    what the flow suites drive. The Phase 1 fixture is 'incoming' and is left alone so the
    ...    two suites do not fight over one record.
    ${id}=    Get Variable Value    ${INTERNAL_OPERATION_TYPE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Location Under Test
    Ensure Stock Destination Location Under Test
    ${name}=    Unique Display Name    Robot Internal Type
    ${code}=    Unique Code    stkint
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'always', 'shipping_policy': 'partial', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INTERNAL_OPERATION_TYPE_ID}    ${id}

Create Stock Transfer
    [Documentation]    Creates a draft transfer and returns its id and etag. The transfer number
    ...    is generated server-side, so none is sent: a client-supplied one is overwritten.
    [Arguments]    ${operation_type_id}
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $operation_type_id, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Add Stock Move
    [Documentation]    Adds one line of demand to a transfer and returns the move id.
    ...    Quantities travel as strings: BR §7.3 forbids floating point for stock quantities, and
    ...    a JSON number would be parsed as a double before it ever reached the decimal column.
    [Arguments]    ${transfer_id}    ${variant_id}    ${quantity}
    ${qty}=    Convert To String    ${quantity}
    ${resp}=    POST On Session    api    ${STOCK_MOVE_API}
    ...    json=${{ {'transfer_id': $transfer_id, 'product_variant_id': $variant_id, 'demand_quantity': $qty, 'base_demand_quantity': $qty, 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Ensure Warehouse Under Test
    [Documentation]    A warehouse and, with it, the locations it creates for itself: a root
    ...    named after its code, and a Stock location beneath. One-step flows both ways, so
    ...    those two are the whole tree until a test reconfigures a flow.
    ${id}=    Get Variable Value    ${WAREHOUSE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Warehouse
    ${code}=    Unique Warehouse Code    W
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'central', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${WAREHOUSE_ID}    ${id}
    Set Global Variable    ${WAREHOUSE_ETAG}    ${etag}
    Set Global Variable    ${WAREHOUSE_CODE}    ${code}

Ensure Secondary Warehouse Under Test
    [Documentation]    A second warehouse, for the rules that need two: a supply route runs
    ...    between different warehouses, and a hierarchy needs somewhere to hang from.
    ${id}=    Get Variable Value    ${SECONDARY_WAREHOUSE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Second Warehouse
    ${code}=    Unique Warehouse Code    S
    ${resp}=    POST On Session    api    ${WAREHOUSE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'warehouse_role': 'pos', 'incoming_flow': 'one_step', 'outgoing_flow': 'one_step', 'status': 'active', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SECONDARY_WAREHOUSE_ID}    ${id}
    Set Global Variable    ${SECONDARY_WAREHOUSE_ETAG}    ${etag}

Ensure Storage Category Under Test
    [Documentation]    A capacity and mixing policy a location may carry.
    ${id}=    Get Variable Value    ${STORAGE_CATEGORY_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Storage Category
    ${code}=    Unique Code    storcat
    ${resp}=    POST On Session    api    ${STORAGE_CATEGORY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'allow_new_item_policy': 'allow', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STORAGE_CATEGORY_ID}    ${id}
    Set Global Variable    ${STORAGE_CATEGORY_ETAG}    ${etag}

Find Warehouse Location By Code
    [Documentation]    Resolves one of a warehouse's own locations. The system-generated ones
    ...    carry the codes the flow topology names, so 'Stock' finds the place goods live.
    [Arguments]    ${warehouse_id}    ${code}
    ${graph}=    Set Variable    {"and":[{"if":["warehouse_id","=","${warehouse_id}"]},{"if":["code","=","${code}"]}]}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    len($items) == 0    RETURN    ${EMPTY}
    RETURN    ${items}[0][id]

Unique Warehouse Code
    [Documentation]    A warehouse code is at most five characters, because it is the first
    ...    segment of every location path beneath it. The shared Unique Code keyword produces
    ...    longer values, so warehouses get their own.
    [Arguments]    ${prefix}
    ${suffix}=    Unique Suffix
    ${code}=    Set Variable    ${prefix}${suffix}
    RETURN    ${code[:5]}

Ensure Supplier Location Under Test
    [Documentation]    The 'vendor' virtual location a receipt draws from. A receipt has to have
    ...    a source, and it is not a place the company holds stock — which is why location_usage
    ...    exists and why only 'internal' counts as company-owned (BR §4.2.2).
    ${id}=    Get Variable Value    ${SUPPLIER_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Supplier
    ${code}=    Unique Code    stksup
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'vendor', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SUPPLIER_LOCATION_ID}    ${id}

Ensure Receipt Operation Type Under Test
    [Documentation]    An incoming type reserving manually and never creating a backorder, so a
    ...    receipt used purely to seed stock cannot leave a stray transfer behind.
    ${id}=    Get Variable Value    ${RECEIPT_OPERATION_TYPE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Receipt Flow Type
    ${code}=    Unique Code    stkrcp
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'operation_code': 'incoming', 'reservation_method': 'manual', 'backorder_policy': 'never', 'shipping_policy': 'partial', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${RECEIPT_OPERATION_TYPE_ID}    ${id}

Ensure Inventory Loss Location Under Test
    [Documentation]    The 'inventory_loss' virtual location an adjustment balances against
    ...    (BR §4.2.7.2). A count that finds more than the books said has to take the difference
    ...    from somewhere, and this is the somewhere.
    ${id}=    Get Variable Value    ${INVENTORY_LOSS_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Inventory Loss
    ${code}=    Unique Code    stkloss
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'inventory_loss', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${INVENTORY_LOSS_LOCATION_ID}    ${id}

Ensure Scrap Location Under Test
    [Documentation]    The 'scrap' virtual location written-off goods move to (BR §4.2.9.1). Stock
    ...    that lands here has left usable inventory without being deleted, which is what makes a
    ...    scrap auditable rather than a disappearance.
    ${id}=    Get Variable Value    ${SCRAP_LOCATION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    ${name}=    Unique Display Name    Robot Scrap Location
    ${code}=    Unique Code    stkscrp
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'location_usage': 'scrap', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SCRAP_LOCATION_ID}    ${id}

Ensure Correction Operation Type Under Test
    [Documentation]    The internal operation type adjustments and scraps generate their movements
    ...    through (decision F3).
    ...
    ...    Its code is the fixed 'INV_CORRECTION', not a Unique Code like the other fixtures: the
    ...    correction helper resolves it by that exact string, and the constant it matches is
    ...    models.StockCorrectionOperationTypeCode. A randomised code here would make every
    ...    adjustment and scrap fail with "no internal operation type", which reads as a data
    ...    problem rather than a fixture mismatch.
    ...
    ...    Because the code is fixed it can already exist from an earlier run, so a conflict is
    ...    tolerated and the existing record reused.
    ${id}=    Get Variable Value    ${CORRECTION_OPERATION_TYPE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Org
    Ensure Inventory Location Under Test
    Ensure Inventory Loss Location Under Test
    ${name}=    Unique Display Name    Robot Correction Type
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    json=${{ {'code': 'INV_CORRECTION', 'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'shipping_policy': 'partial', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    IF    ${resp.status_code} == 200 or ${resp.status_code} == 201
        ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ELSE
        ${id}=    Find Operation Type By Code    INV_CORRECTION
    END
    Set Global Variable    ${CORRECTION_OPERATION_TYPE_ID}    ${id}

Find Operation Type By Code
    [Documentation]    Reads back an operation type by its code, for the fixed-code fixture above.
    [Arguments]    ${code}
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('code') == $code    RETURN    ${item}[id]
    END
    Fail    No operation type with code '${code}' exists; the correction fixture cannot continue

Ensure Correction Fixtures
    [Documentation]    Everything an adjustment or a scrap needs before it can generate a movement.
    Ensure Correction Operation Type Under Test
    Ensure Inventory Loss Location Under Test
    Ensure Scrap Location Under Test

Read Stock Quant Id
    [Documentation]    Finds the balance row for a variant in a location, which the count actions
    ...    address by id. Fails loudly when absent: every caller here has just received stock, so a
    ...    missing balance means the receipt did not land and the test should say so.
    [Arguments]    ${variant_id}    ${location_id}
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('product_variant_id') == $variant_id and $item.get('location_id') == $location_id
            RETURN    ${item}[id]
        END
    END
    Fail    No stock balance for variant '${variant_id}' in location '${location_id}'

Read Stock Quant Field
    [Documentation]    Reads one field off a balance, for asserting count metadata.
    [Arguments]    ${quant_id}    ${field}
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}/${quant_id}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[data].get('${field}')

Receive Stock Into Location
    [Documentation]    Puts a quantity of a variant into a location by running a real receipt:
    ...    create -> add move -> confirm -> validate.
    ...
    ...    It goes the long way round on purpose. The quant is read-only to clients (AC-STOCK-002),
    ...    so there is no shortcut that seeds a balance, and a test fixture that had one would be
    ...    proving something the product does not allow.
    ...
    ...    A receipt draws from the supplier location, where nothing is on hand, so its moves
    ...    reserve nothing and validate ships the full demand from an unreserved source.
    [Arguments]    ${variant_id}    ${location_id}    ${quantity}
    Ensure Supplier Location Under Test
    Ensure Receipt Operation Type Under Test
    ${qty}=    Convert To String    ${quantity}

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $RECEIPT_OPERATION_TYPE_ID, 'source_location_id': $SUPPLIER_LOCATION_ID, 'destination_location_id': $location_id, 'org_id': $INV_ORG_ID} }}
    ${transfer_id}    ${etag}=    Response Should Be Create Success    ${resp}

    ${resp}=    POST On Session    api    ${STOCK_MOVE_API}
    ...    json=${{ {'transfer_id': $transfer_id, 'product_variant_id': $variant_id, 'demand_quantity': $qty, 'base_demand_quantity': $qty, 'source_location_id': $SUPPLIER_LOCATION_ID, 'destination_location_id': $location_id, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ${move_id}    ${move_etag}=    Response Should Be Create Success    ${resp}

    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/confirm
    ...    json=${{ {} }}    expected_status=any
    RETURN    ${transfer_id}    ${move_id}

Read Stock On Hand
    [Documentation]    Sums on_hand_quantity across every balance of a variant in a location.
    ...    Returns 0 when no balance exists yet, which is the state before any receipt: absent and
    ...    zero mean the same thing to a caller asking how much is there.
    [Arguments]    ${variant_id}    ${location_id}
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${total}=    Set Variable    ${0}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('product_variant_id') == $variant_id and $item.get('location_id') == $location_id
            ${qty}=    Convert To Number    ${item.get('on_hand_quantity') or 0}
            ${total}=    Evaluate    ${total} + ${qty}
        END
    END
    RETURN    ${total}

Read Stock Reserved
    [Documentation]    Sums reserved_quantity across every balance of a variant in a location.
    [Arguments]    ${variant_id}    ${location_id}
    ${resp}=    GET On Session    api    ${STOCK_QUANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${total}=    Set Variable    ${0}
    FOR    ${item}    IN    @{resp.json()}[items]
        IF    $item.get('product_variant_id') == $variant_id and $item.get('location_id') == $location_id
            ${qty}=    Convert To Number    ${item.get('reserved_quantity') or 0}
            ${total}=    Evaluate    ${total} + ${qty}
        END
    END
    RETURN    ${total}

Delete Inventory Seed Data
    [Documentation]    Best-effort removal of the fixtures the Inventory suites created on
    ...    the fly, referencing-before-referenced: variants and attribute values first,
    ...    then their owners, then the master data every template points at. Every FK in
    ...    this module is ON DELETE NO ACTION apart from the junctions, so a wrong order
    ...    leaves rows behind rather than cascading.
    #    Movement unwinds innermost-first: a move line points at its move and its transfer, a
    #    move points at its transfer, and the transfer points at the operation type and both
    #    locations. Deleting the transfer first would leave its moves orphaned rather than
    #    cascading, since every FK here is ON DELETE NO ACTION.
    Delete Inventory Fixture    ${STOCK_MOVE_LINE_API}    STOCK_MOVE_LINE_ID
    Delete Inventory Fixture    ${STOCK_MOVE_API}    STOCK_MOVE_ID
    #    A backorder points at the transfer it carries the remainder of, so it goes first.
    Delete Inventory Fixture    ${STOCK_TRANSFER_API}    BACKORDER_TRANSFER_ID
    Delete Inventory Fixture    ${STOCK_TRANSFER_API}    STOCK_TRANSFER_ID
    #    The operation type points at the location, and a quant points at both the location
    #    and a variant, so stock unwinds before the products it references.
    Delete Inventory Fixture    ${STOCK_OPERATION_TYPE_API}    INTERNAL_OPERATION_TYPE_ID
    Delete Inventory Fixture    ${STOCK_OPERATION_TYPE_API}    STOCK_OPERATION_TYPE_ID
    #    Warehouse configuration points at warehouses and locations, so it unwinds before both.
    #    A putaway rule names two locations and a warehouse; a supply relation names two
    #    warehouses.
    Delete Inventory Fixture    ${PUTAWAY_RULE_API}    PUTAWAY_RULE_ID
    Delete Inventory Fixture    ${SUPPLY_RELATION_API}    SUPPLY_RELATION_ID
    #    The child points at the location under test, so it must go first.
    Delete Inventory Fixture    ${INVENTORY_LOCATION_API}    CHILD_INVENTORY_LOCATION_ID
    Delete Inventory Fixture    ${INVENTORY_LOCATION_API}    STOCK_DEST_LOCATION_ID
    Delete Inventory Fixture    ${INVENTORY_LOCATION_API}    INVENTORY_LOCATION_ID
    #    A warehouse creates its own locations, which point back at it, so those go first. They
    #    are system-generated and refuse a normal delete while the warehouse is live, which is
    #    why the warehouse is archived rather than deleted here.
    Archive Inventory Fixture    ${WAREHOUSE_API}    SECONDARY_WAREHOUSE_ID
    Archive Inventory Fixture    ${WAREHOUSE_API}    WAREHOUSE_ID
    #    A location may carry a storage category, so the category outlives the locations.
    Delete Inventory Fixture    ${STORAGE_CATEGORY_API}    STORAGE_CATEGORY_ID
    Delete Matrix Template Fixture
    Delete Inventory Fixture    ${PRODUCT_VARIANT_API}    PRODUCT_VARIANT_ID
    Delete Inventory Fixture    ${PRODUCT_TEMPLATE_API}    PRODUCT_TEMPLATE_ID
    Delete Inventory Fixture    ${ATTRIBUTE_VALUE_API}    ATTRIBUTE_VALUE_ID
    Delete Inventory Fixture    ${PRODUCT_ATTRIBUTE_API}    PRODUCT_ATTRIBUTE_ID
    Delete Inventory Fixture    ${BRAND_API}    BRAND_ID
    # The child points at the category under test, so it must go first.
    Delete Inventory Fixture    ${PRODUCT_CATEGORY_API}    CHILD_CATEGORY_ID
    Delete Inventory Fixture    ${PRODUCT_CATEGORY_API}    PRODUCT_CATEGORY_ID
    Delete Inventory Fixture    ${PRODUCT_TYPE_API}    PRODUCT_TYPE_ID
