*** Settings ***
Documentation     Reading a single Unit of Measure, including the `category` edge the
...               listing resolves category names through.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Under Test
Test Tags         essential    uom    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    Set Global Variable    ${UOM_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}
    ...    params=${{ {'fields': ['symbol', 'factor', 'rounding']} }}
    Response Status Should Be    ${resp}    200

Get With Edge Columns Succeeds
    [Documentation]    Resolving category.name is how a UoM row shows its category as a
    ...    label instead of a ULID.
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}
    ...    params=${{ {'fields': ['symbol', 'category.name']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}
    ...    params=${{ {'fields': ['symbol', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOM_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOM_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
