*** Settings ***
Documentation     Existence checks over Product Attribute Values.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Attribute Value Under Test
Test Tags         inventory    product_attribute_value    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}/exists
    ...    json=${{ {'org_id': $INV_ORG_ID, 'ids': [$ATTRIBUTE_VALUE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Several Ids Succeeds
    [Documentation]    No bulk-seed keyword exists for attribute values, so a handful are
    ...    created inline and combined with fake ids to exercise the real/fake split.
    ${id_a}    ${etag_a}=    Create Attribute Value    ${PRODUCT_ATTRIBUTE_ID}    Robot Exists A
    ${id_b}    ${etag_b}=    Create Attribute Value    ${PRODUCT_ATTRIBUTE_ID}    Robot Exists B
    ${id_c}    ${etag_c}=    Create Attribute Value    ${PRODUCT_ATTRIBUTE_ID}    Robot Exists C
    ${fakes}=    Not Found Id List    3
    ${ids}=    Combine Lists    ${{ [$id_a, $id_b, $id_c] }}    ${fakes}
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}/exists
    ...    json=${{ {'org_id': $INV_ORG_ID, 'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=3    not_existing=3
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id_a}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id_b}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id_c}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}/exists
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}/exists
    ...    json=${{ {'org_id': $INV_ORG_ID, 'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
