*** Settings ***
Documentation     Existence checks over Vendor Profiles. Like the relationship suite there is
...               no bulk seed: a profile is one-per-party, so seeding fifty would mean seeding
...               fifty parties to hang them off. The many-ids case uses the one real profile
...               plus known-absent ids, which exercises the same partition of the response.
Library           Collections
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/exists
    ...    json=${{ {'ids': [$VENDOR_PROFILE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Mixed Ids Succeeds
    ${fakes}=    Not Found Id List    5
    ${ids}=    Create List    ${VENDOR_PROFILE_ID}
    ${ids}=    Combine Lists    ${ids}    ${fakes}
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/exists
    ...    json=${{ {'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=5

Exists With Missing Required Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/exists
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    ids

Exists With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}/exists
    ...    json=${{ {'ids': ['invalid']} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    ids
