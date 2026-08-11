*** Settings ***
Documentation     Reading a single Product Template, including the `variants` edge the
...               template detail page lists its SKUs through.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_template    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'fields': ['name', 'status', 'category_id']} }}
    Response Status Should Be    ${resp}    200

Get With Edge Columns Succeeds
    [Documentation]    The `variants` edge is what the template detail page's related-records
    ...    section reads; a template with no variants still answers, with an empty list.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'fields': ['name', 'variants.sku']} }}
    Response Status Should Be    ${resp}    200

Get With Classification Edges Succeeds
    [Documentation]    A template renders its type, category and brand by name rather than by
    ...    ULID, which is what these three many:one edges are for.
    Ensure Product Type Under Test
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'fields': ['name', 'product_type.code', 'category.code']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
