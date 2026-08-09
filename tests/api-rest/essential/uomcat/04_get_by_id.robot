*** Settings ***
Documentation     Reading a single UoM Category, including the `uoms` edge that the
...               category detail page lists its members through.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Category Under Test
Test Tags         essential    uomcat    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uomcat.json    200
    Set Global Variable    ${UOMCAT_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    params=${{ {'fields': ['name', 'reference_uom_id']} }}
    Response Status Should Be    ${resp}    200

Get With Edge Columns Succeeds
    [Documentation]    The `uoms` edge is what the category detail page's related-records
    ...    section reads; a category with no members still answers, with an empty list.
    Ensure Reference Uom
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    params=${{ {'fields': ['name', 'uoms.symbol']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOMCAT_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
