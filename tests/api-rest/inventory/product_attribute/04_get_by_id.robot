*** Settings ***
Documentation     Reading a single Product Attribute. The model declares no edges_from for
...               its values — attribute_value only declares a many:one edges_to back at the
...               attribute — so there is no values edge projection to exercise here.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute.json    200
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    params=${{ {'fields': ['code', 'name', 'data_type']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${PRODUCT_ATTRIBUTE_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
