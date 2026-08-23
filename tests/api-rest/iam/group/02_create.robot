*** Settings ***
Documentation     Creating a group, with the LangJson shape of "name" as the focus: an
...               object keyed by BCP47 language code, never a bare string.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seed User
Test Tags         iam    group    create


*** Variables ***
${GROUP_SCHEMA}    ${IAM_SCHEMA_DIR}/group.json


*** Test Cases ***
Create Group With Both Languages Succeeds
    ${suffix}=    Unique Suffix
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': 'Robot Group ' + $suffix, 'vi-VN': 'Nhom Robot ' + $suffix}, 'description': {'en-US': 'Created by the API test suite'}, 'owner_id': $SEED_USER_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Suite Variable    ${CREATED_GROUP_ID}    ${id}

Created Group Round Trips Both Languages
    [Documentation]    A read returns the whole document rather than one language: the
    ...    server localizes what it filters and sorts on, not what it returns, and the
    ...    client picks the key it needs.
    ${resp}=    GET On Session    api    ${GROUP_API}/${CREATED_GROUP_ID}
    Response Status Should Be    ${resp}    200
    ${name}=    Evaluate    $resp.json()['item']['name']
    Dictionary Should Contain Key    ${name}    en-US
    Dictionary Should Contain Key    ${name}    vi-VN

Create Group With One Language Succeeds
    [Documentation]    A langjson column does not require every supported language.
    ${suffix}=    Unique Suffix
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': 'English Only Group ' + $suffix}, 'owner_id': $SEED_USER_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${GROUP_API}/${id}    expected_status=any

Create Group With Bare String Name Fails
    [Documentation]    A langjson column needs the language envelope; a bare string names
    ...    no language and so cannot be stored as one.
    [Tags]    negative
    ${suffix}=    Unique Suffix
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': 'Bare String ' + $suffix, 'owner_id': $SEED_USER_ID} }}
    ...    expected_status=any
    Should Be True    ${{ $resp.status_code >= 400 }}
    ...    msg=A bare string should not be accepted for a langjson column, got ${resp.status_code}

Delete Created Group
    [Documentation]    Runs last so the group above exists for the read assertions.
    [Tags]    cleanup
    DELETE On Session    api    ${GROUP_API}/${CREATED_GROUP_ID}    expected_status=any
