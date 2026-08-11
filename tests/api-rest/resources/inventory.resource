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

Delete Inventory Seed Data
    [Documentation]    Best-effort removal of the fixtures the Inventory suites created on
    ...    the fly, referencing-before-referenced: variants and attribute values first,
    ...    then their owners, then the master data every template points at. Every FK in
    ...    this module is ON DELETE NO ACTION apart from the junctions, so a wrong order
    ...    leaves rows behind rather than cascading.
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
