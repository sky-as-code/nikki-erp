*** Settings ***
Documentation     Existence checks over Units of Measure. The sample units are seeded
...               once per execution and shared with 07_search.robot.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Uom Under Test
...               AND    Ensure Seeded Uoms    50
Test Tags         essential    uom    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${UOM_API}/exists
    ...    json=${{ {'ids': [$UOM_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Many Ids Succeeds
    [Tags]    seed
    ${existing}=    Get Slice From List    ${SEEDED_UOM_IDS}    0    45
    ${fakes}=    Not Found Id List    5
    ${ids}=    Combine Lists    ${existing}    ${fakes}
    ${resp}=    POST On Session    api    ${UOM_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=45    not_existing=5

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
