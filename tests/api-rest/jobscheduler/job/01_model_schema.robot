*** Settings ***
Documentation     Job Scheduler/Job - Get model schema.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Test Tags         jobscheduler    job    schema


*** Test Cases ***
Get Job Model Schema
    ${resp}=    GET On Session    api    ${JOB_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Job Schema Declares No Timezone Field
    [Documentation]    The scheduler is UTC-only by construction. A per-job timezone would
    ...    reintroduce exactly the ambiguity that choice removes, so its absence is part of
    ...    the published contract rather than an omission.
    ${resp}=    GET On Session    api    ${JOB_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json()['fields'].keys())
    Should Not Contain    ${names}    timezone

Job Schema Declares No Archive Field
    [Documentation]    Guards the same decision 08_delete.robot asserts over HTTP: a job is
    ...    deleted permanently, never archived.
    ${resp}=    GET On Session    api    ${JOB_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json()['fields'].keys())
    Should Not Contain    ${names}    is_archived
