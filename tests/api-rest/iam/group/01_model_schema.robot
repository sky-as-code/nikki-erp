*** Settings ***
Documentation     The group model schema, and the declaration that makes the rest of this
...               suite meaningful: "name" must be a langjson field, because everything the
...               search file asserts about locale-aware ordering follows from that.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Test Tags         iam    group    schema


*** Test Cases ***
Get Group Model Schema
    ${resp}=    GET On Session    api    ${GROUP_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Group Name Is A Lang Json Field
    [Documentation]    Pins the premise of the search suite. If "name" ever became a plain
    ...    string, the locale tests would still pass while testing nothing, so the type is
    ...    asserted here rather than assumed there.
    ${resp}=    GET On Session    api    ${GROUP_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${type}=    Evaluate
    ...    $resp.json()['fields']['name']['data_type']['name']
    Should Be Equal    ${type}    nikkiLangJson
    ...    msg=iam_group.name must be a langjson column for the locale search tests to mean anything
