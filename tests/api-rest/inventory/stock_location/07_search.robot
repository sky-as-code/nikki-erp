*** Settings ***
Documentation     Searching Stock Locations. A location is org-scoped, so every search
...               carries org_id — omitting it returns nothing rather than everything.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Location Under Test
Test Tags         inventory    stock_location    search


*** Variables ***
${STOCK_LOCATION_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/stock_location.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${STOCK_LOCATION_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 0, 'size': 5} }}
    Response Should Be Search Success    ${resp}    ${STOCK_LOCATION_SCHEMA}    size=5    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${STOCK_LOCATION_SCHEMA}    size=50    page=99    item_count=0

Search By Code Succeeds
    ${graph}=    Set Variable    {"if":["code", "=", "${STOCK_LOCATION_CODE}"]}
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${STOCK_LOCATION_SCHEMA}    size=50    page=0    item_count=1

Search By Location Type Succeeds
    [Documentation]    Filtering to `internal` is how a caller lists the locations that
    ...    actually hold company-owned stock, so it is the filter the balance views rely on.
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["location_type", "=", "internal"]}'} }}
    Response Should Be Search Success    ${resp}    ${STOCK_LOCATION_SCHEMA}    size=50    page=0

Search With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${STOCK_LOCATION_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
