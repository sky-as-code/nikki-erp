*** Settings ***
Documentation     Job Scheduler/Job - Get by id.
Resource          resources/jobscheduler.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Job Under Test
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    get


*** Test Cases ***
Get Job By Id Succeeds
    ${resp}=    GET On Session    api    ${JOB_API}/${JOB_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${JOBSCHEDULER_SCHEMA_DIR}/job.json    200
    Should Be Equal    ${item}[id]    ${JOB_ID}
    Should Be Equal    ${item}[module_name]    ${ROBOT_MODULE}

Get Job Returns The Action Config It Was Registered With
    [Documentation]    action_config is what the executor runs. A job whose stored config
    ...    differs from the one submitted would do something other than what was registered.
    ${id}    ${etag}    ${payload}=    Create Job
    ${resp}=    GET On Session    api    ${JOB_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${JOBSCHEDULER_SCHEMA_DIR}/job.json    200
    Should Be Equal    ${item}[action_config][url]    ${payload}[action_config][url]
    Should Be Equal    ${item}[action_config][method]    ${payload}[action_config][method]

Get Job With Selected Fields Returns Only Those
    ${resp}=    GET On Session    api    ${JOB_API}/${JOB_ID}
    ...    params=${{ {'fields': ['name', 'cron_expression']} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[item]
    Dictionary Should Contain Key    ${item}    name
    Dictionary Should Contain Key    ${item}    cron_expression
    Dictionary Should Not Contain Key    ${item}    action_config

Get Unknown Job Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${JOB_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get Job With Malformed Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${JOB_API}/not-a-ulid    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
