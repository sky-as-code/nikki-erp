*** Settings ***
Documentation     Existence checks over Relationships. Unlike party and comm channel there is
...               no bulk seed: a relationship needs two distinct parties, so seeding fifty
...               would mean seeding a hundred parties to join. The many-ids case uses the
...               one real relationship plus known-absent ids, which exercises the same
...               partition of the response.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Relationship Under Test
Test Tags         contacts    relationship    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/exists
    ...    json=${{ {'ids': [$RELATIONSHIP_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Mixed Ids Succeeds
    ${fakes}=    Not Found Id List    5
    ${ids}=    Create List    ${RELATIONSHIP_ID}
    ${ids}=    Combine Lists    ${ids}    ${fakes}
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=5

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
