*** Settings ***
Documentation     Reading a single Product Attribute Value.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Attribute Value Under Test
Test Tags         inventory    product_attribute_value    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute_value.json    200
    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    params=${{ {'fields': ['code', 'name', 'price_extra']} }}
    Response Status Should Be    ${resp}    200

Get With Edge Columns Succeeds
    [Documentation]    The `attribute` edge is what the value detail page names its owning
    ...    attribute through.
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    params=${{ {'fields': ['code', 'attribute.code']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${ATTRIBUTE_VALUE_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
