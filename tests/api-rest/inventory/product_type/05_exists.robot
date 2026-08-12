*** Settings ***
Documentation     Existence checks over Product Types.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Type Under Test
Test Tags         inventory    product_type    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/exists
    ...    json=${{ {'ids': [$PRODUCT_TYPE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Mixed Ids Succeeds
    [Documentation]    Real and fake ids in one call, so the split is exercised rather than
    ...    an all-present or all-absent shortcut.
    ${fakes}=    Not Found Id List    3
    ${ids}=    Combine Lists    ${{ [$PRODUCT_TYPE_ID] }}    ${fakes}
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=3

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
