*** Settings ***
Documentation     Searching Units of Measure. The category filter is the one the UoM
...               Category detail page uses to list a category's members, so it is a UI
...               contract as much as an API one.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seeded Uoms    50
Test Tags         essential    uom    search


*** Variables ***
${UOM_SCHEMA}    ${ESSENTIAL_SCHEMA_DIR}/uom.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=50    page=99    item_count=0

Search By Category Succeeds
    [Documentation]    The exact filter the category detail page issues for its
    ...    related-records table. category_id is a plain FK, so it filters on equality
    ...    rather than the `linked` operator the IAM assignment pages use.
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'graph': '{"if":["category_id", "=", "' + $UOMCAT_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=50    page=0
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=The seeded units belong to this category
    FOR    ${item}    IN    @{items}
        Should Be Equal    ${item}[category_id]    ${UOMCAT_ID}
        ...    msg=The category filter leaked a UoM from another category
    END

Search By Uom Type Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'graph': '{"if":["uom_type", "=", "bigger_equal"]}'} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=50    page=0

Search By Name Succeeds
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'graph': '{"if":["name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${UOM_SCHEMA}    size=50    page=0

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOM_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'fields': ['symbol', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
