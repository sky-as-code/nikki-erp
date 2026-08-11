*** Settings ***
Documentation     Reading a single Product Category.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Category Under Test
Test Tags         inventory    product_category    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_category.json    200
    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    params=${{ {'fields': ['code', 'name', 'parent_category_id']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
