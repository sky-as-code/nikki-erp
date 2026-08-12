*** Settings ***
Documentation     Reading a single Stock Location.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Location Under Test
Test Tags         inventory    stock_location    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/stock_location.json    200
    Set Global Variable    ${STOCK_LOCATION_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    params=${{ {'fields': ['code', 'name', 'location_type']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/${STOCK_LOCATION_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
