*** Settings ***
Documentation     Searching Product Types. Unlike every other Inventory resource a product
...               type is global rather than org-scoped, so its searches carry no org_id.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Type Under Test
Test Tags         inventory    product_type    search


*** Variables ***
${PRODUCT_TYPE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_type.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}
    Response Should Be Search Success    ${resp}    ${PRODUCT_TYPE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}
    ...    params=${{ {'page': 0, 'size': 3} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_TYPE_SCHEMA}    size=3    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}    params=${{ {'page': 99} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_TYPE_SCHEMA}    size=50    page=99    item_count=0

Search By Code Succeeds
    [Documentation]    Filters on the type under test's own code, so the result is
    ...    deterministic on any database rather than depending on seeded content.
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}
    ...    params=${{ {'graph': '{"if":["code", "=", "' + $PRODUCT_TYPE_CODE + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_TYPE_SCHEMA}    size=50    page=0    item_count=1

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}
    ...    params=${{ {'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
