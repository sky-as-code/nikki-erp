*** Settings ***
Documentation     Searching Stock Operation Types. The resource is org-scoped, so every
...               search carries org_id — omitting it returns nothing rather than everything.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    search


*** Variables ***
${STOCK_OPERATION_TYPE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${STOCK_OPERATION_TYPE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 0, 'size': 5} }}
    Response Should Be Search Success    ${resp}    ${STOCK_OPERATION_TYPE_SCHEMA}    size=5    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${STOCK_OPERATION_TYPE_SCHEMA}
    ...    size=50    page=99    item_count=0

Search By Code Succeeds
    ${graph}=    Set Variable    {"if":["code", "=", "${STOCK_OPERATION_TYPE_CODE}"]}
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${STOCK_OPERATION_TYPE_SCHEMA}
    ...    size=50    page=0    item_count=1

Search By Operation Code Succeeds
    [Documentation]    Listing the receipt types is how a transfer screen offers the ones
    ...    valid for an inbound movement.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["operation_code", "=", "incoming"]}'} }}
    Response Should Be Search Success    ${resp}    ${STOCK_OPERATION_TYPE_SCHEMA}    size=50    page=0

Search With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
