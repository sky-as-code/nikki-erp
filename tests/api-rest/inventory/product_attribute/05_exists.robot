*** Settings ***
Documentation     Existence checks over Product Attributes.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}/exists
    ...    json=${{ {'ids': [$PRODUCT_ATTRIBUTE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Several Ids Succeeds
    [Documentation]    A handful of real ids alongside fakes, so the split is exercised
    ...    rather than an all-present or all-absent shortcut.
    ${ids}=    Create List    ${PRODUCT_ATTRIBUTE_ID}
    FOR    ${index}    IN RANGE    3
        ${name}=    Unique Display Name    Robot Exists Attribute ${index}
        ${code}=    Unique Code    exists${index}
        ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    ${fakes}=    Not Found Id List    3
    ${all_ids}=    Combine Lists    ${ids}    ${fakes}
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}/exists
    ...    json=${{ {'ids': $all_ids} }}
    Response Should Be Exists Success    ${resp}    existing=4    not_existing=3

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
