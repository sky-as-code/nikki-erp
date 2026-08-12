*** Settings ***
Documentation     Existence checks over Product Templates. The sample templates are seeded
...               once per execution and shared with 07_search.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Product Template Under Test
...               AND    Ensure Seeded Product Templates    50
Test Tags         inventory    product_template    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/exists
    ...    json=${{ {'ids': [$PRODUCT_TEMPLATE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Many Ids Succeeds
    [Tags]    seed
    ${existing}=    Get Slice From List    ${SEEDED_TEMPLATE_IDS}    0    45
    ${fakes}=    Not Found Id List    5
    ${ids}=    Combine Lists    ${existing}    ${fakes}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=45    not_existing=5

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
