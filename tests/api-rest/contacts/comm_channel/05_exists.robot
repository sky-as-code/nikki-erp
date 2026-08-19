*** Settings ***
Documentation     Existence checks over Comm Channels. The sample channels are seeded once
...               per execution and shared with 07_search.robot.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
...               AND    Ensure Seeded Comm Channels    50
Test Tags         contacts    comm_channel    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/exists
    ...    json=${{ {'ids': [$COMM_CHANNEL_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Many Ids Succeeds
    [Tags]    seed
    ${existing}=    Get Slice From List    ${SEEDED_COMM_CHANNEL_IDS}    0    45
    ${fakes}=    Not Found Id List    5
    ${ids}=    Combine Lists    ${existing}    ${fakes}
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=45    not_existing=5

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
