*** Settings ***
Documentation     Bruno: IAM/User/User - Get by ID. The edge-column tests use the seed
...               user/group (created on the fly when the environment provides none).
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    Set Global Variable    ${USER_ETAG}    ${item}[etag]

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ...    params=${{ {'fields': ['email', 'display_name']} }}
    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user_columns.json    200

Get With Edge All Columns Succeeds
    [Tags]    seed
    Ensure Seed Group
    ${resp}=    GET On Session    api    ${USER_API}/${SEED_USER_ID}
    ...    params=${{ {'fields': ['email', 'groups']} }}
    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user_edge_all_columns.json    200

Get With Edge Columns Succeeds
    [Tags]    seed
    Ensure Seed Group
    ${resp}=    GET On Session    api    ${USER_API}/${SEED_USER_ID}
    ...    params=${{ {'fields': ['email', 'groups.name']} }}
    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user_edge_columns.json    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ...    params=${{ {'fields': ['display_name', 'email', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}/not-valid-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
