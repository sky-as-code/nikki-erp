*** Settings ***
Documentation     Searching Product Attributes. org_id is required-for-create, so every
...               search here is org-scoped.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    search


*** Variables ***
${PRODUCT_ATTRIBUTE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_attribute.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_ATTRIBUTE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 0, 'size': 3} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_ATTRIBUTE_SCHEMA}    size=3    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_ATTRIBUTE_SCHEMA}    size=50    page=99    item_count=0

Search By Code Succeeds
    [Documentation]    Filters on the attribute under test's own code, so the result is
    ...    deterministic on any database rather than depending on seeded content.
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["code", "=", "' + $PRODUCT_ATTRIBUTE_CODE + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${PRODUCT_ATTRIBUTE_SCHEMA}    size=50    page=0    item_count=1

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
