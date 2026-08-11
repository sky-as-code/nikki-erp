*** Settings ***
Documentation     Reading a single Product Variant, including the effective-product endpoint
...               that flattens the template and variant into the one shape a consumer reads
...               (BR §7.5, AC-PROD-032).
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_variant    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=${{ {'fields': ['sku', 'combination_key', 'status']} }}
    Response Status Should Be    ${resp}    200

Get With Edge Columns Succeeds
    [Documentation]    The `template` edge is how the variant detail page names its parent
    ...    product line instead of showing a ULID.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=${{ {'fields': ['sku', 'template.name']} }}
    Response Status Should Be    ${resp}    200

Get Effective Product Succeeds
    [Documentation]    BR §7.5 / AC-PROD-032: the flattened product a consumer reads instead
    ...    of joining the two halves itself. Re-implementing inheritance in each consumer is
    ...    exactly what this endpoint exists to prevent. is_selectable is served rather than
    ...    derived, so no caller has to re-apply the archive and status rules.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/effective
    Response Status Should Be    ${resp}    200
    ${data}=    Set Variable    ${resp.json()}
    Validate Json Schema    ${data}    ${INVENTORY_SCHEMA_DIR}/effective_product.json
    Should Be Equal    ${data}[variant_id]    ${PRODUCT_VARIANT_ID}
    Should Be Equal    ${data}[template_id]    ${PRODUCT_TEMPLATE_ID}
    Should Not Be Empty    ${data}[display_name]
    ...    msg=display_name is computed from the template name plus attribute values (BR 5.5)

Get Effective Product With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}/effective    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=${{ {'fields': ['sku', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
