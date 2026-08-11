*** Settings ***
Documentation     Searching Product Attribute Values. Like every Inventory resource except
...               product_type, a value is org-scoped (org_id required-for-create), so every
...               search here carries org_id. The "graph" query values are JSON strings that
...               requests URL-encodes via the params dict.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Attribute Value Under Test
Test Tags         inventory    product_attribute_value    search


*** Variables ***
${ATTRIBUTE_VALUE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_attribute_value.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${ATTRIBUTE_VALUE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 0, 'size': 3} }}
    Response Should Be Search Success    ${resp}    ${ATTRIBUTE_VALUE_SCHEMA}    size=3    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${ATTRIBUTE_VALUE_SCHEMA}    size=50    page=99    item_count=0

Search By Attribute Succeeds
    [Documentation]    The query the template's allowed-values picker actually issues: values
    ...    scoped to the owning attribute, filtered via graph rather than a dedicated param.
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["attribute_id", "=", "' + $PRODUCT_ATTRIBUTE_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${ATTRIBUTE_VALUE_SCHEMA}    size=50    page=0    item_count=1

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
