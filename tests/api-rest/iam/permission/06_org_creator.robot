*** Settings ***
Documentation     Creating an organization makes its creator a member, in the same
...               transaction as the create. Membership is what an org-scoped grant and the
...               org switcher both read, so a creator who was not a member could not work
...               in the org they had just made.
Resource          ../../resources/iam.resource
Suite Setup       Create Authorized API Session


*** Test Cases ***
The Creator Of An Org Sees It In Their Context
    [Documentation]    No manage_users call in between: the create alone must do it.
    [Tags]    org    membership
    ${name}=    Unique Display Name    Perm Creator Org
    ${slug}=    Evaluate    'perm-creator-org-' + str(__import__('uuid').uuid4())[:8]
    ${resp}=    POST On Session    api    /v1/iam/iam_org
    ...    json=${{ {'display_name': {'en-US': $name}, 'slug': $slug, 'legal_name': {'en-US': $name}} }}
    ${org_id}    ${etag}=    Response Should Be Create Success    ${resp}
    # The context lists only ready orgs, and an org is ready once every module has set it up.
    Wait Until Org Is Ready    ${org_id}
    ${me}=    GET On Session    api    /v1/iam/me/context
    Response Status Should Be    ${me}    200
    ${org_ids}=    Evaluate    [org['id'] for org in body['orgs']]    namespace=${{ {'body': $me.json()} }}
    Should Contain    ${org_ids}    ${org_id}    msg=The creator must be a member of the org they created
    Should Contain    ${me.json()}[user_org_ids]    ${org_id}
    [Teardown]    DELETE On Session    api    /v1/iam/iam_org/${org_id}    expected_status=any
