*** Settings ***
Documentation     Job Scheduler/Execution - Get model schema.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Test Tags         jobscheduler    execution    schema


*** Test Cases ***
Get Execution Model Schema
    ${resp}=    GET On Session    api    ${EXECUTION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Get Attempt Model Schema
    [Documentation]    The attempt has a schema route but no listing of its own: attempts
    ...    are read as part of an execution, because "why did this end the way it did" is
    ...    not answerable from an attempt in isolation.
    ${resp}=    GET On Session    api    ${ATTEMPT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Execution Schema Carries The Frozen Job Snapshot
    [Documentation]    job_snapshot is what makes history readable after its job is edited
    ...    or deleted. Without it a finished execution could only be described in terms of a
    ...    job that may no longer exist, or may since have changed.
    ${resp}=    GET On Session    api    ${EXECUTION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json()['fields'].keys())
    Should Contain    ${names}    job_snapshot
    Should Contain    ${names}    next_occurrence_at
