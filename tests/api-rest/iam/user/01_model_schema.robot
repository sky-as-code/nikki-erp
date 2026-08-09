*** Settings ***
Documentation     Bruno: IAM/User/User - Get model schema.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Test Tags         iam    user    schema


*** Test Cases ***
Get User Model Schema
    ${resp}=    GET On Session    api    ${USER_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200
