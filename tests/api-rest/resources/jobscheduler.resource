*** Settings ***
Documentation     Job Scheduler helpers shared by the jobscheduler suites: the lazily
...               created job-under-test, bulk sample data for Search, and the
...               registration payload builder. All keywords expect an authorized
...               session with alias "api".
...
...               Note there is no archive helper. None of the three schemas is
...               archivable and 0002002_jobscheduler_iam.sql seeds no set_archived
...               action, so an archive suite would be testing an endpoint that by
...               design does not exist.
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${JOB_API}                  /v1/jobscheduler/jobs
${EXECUTION_API}            /v1/jobscheduler/executions
${ATTEMPT_API}              /v1/jobscheduler/attempts
${JOBSCHEDULER_SCHEMA_DIR}  ${CURDIR}/schemas/jobscheduler

# The module every suite registers its jobs under. Alphanumeric only: module_name is
# compared for exact equality by the delete-by-module endpoint, and the column's regex
# rejects anything else.
${ROBOT_MODULE}             robotjobscheduler

# Ids of jobs created on the fly, removed by Delete Jobscheduler Seed Data.
${SEEDED_JOB_IDS}           ${None}


*** Keywords ***
New Job Payload
    [Documentation]    A valid registration with every required field present.
    ...
    ...    The action is a rest_api one rather than command_bus on purpose: a command_bus
    ...    config is validated against the live CQRS registry, so the payload would have to
    ...    name a command some other module happens to have registered, and the fixture
    ...    would start failing whenever that module changed.
    [Arguments]    ${job_key}=${EMPTY}    ${cron}=*/15 * * * *    &{overrides}
    ${suffix}=    Unique Suffix
    ${key}=    Set Variable If    $job_key    ${job_key}    robot_job_${suffix}
    ${name}=    Unique Display Name    Robot Scheduled Job
    ${payload}=    Create Dictionary
    ...    name=${name}
    ...    job_type=technical
    ...    module_name=${ROBOT_MODULE}
    ...    job_key=${key}
    ...    action_type=rest_api
    ...    action_config=${{ {'method': 'POST', 'url': 'https://api.nikkierp.com/v1/health'} }}
    ...    cron_expression=${cron}
    Set To Dictionary    ${payload}    &{overrides}
    RETURN    ${payload}

Create Job
    [Documentation]    Registers a job and remembers it for the suite teardown.
    [Arguments]    &{overrides}
    ${payload}=    New Job Payload    &{overrides}
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Remember Seeded Job    ${id}
    RETURN    ${id}    ${etag}    ${payload}

Remember Seeded Job
    [Arguments]    ${id}
    ${ids}=    Get Variable Value    ${SEEDED_JOB_IDS}    ${None}
    IF    $ids is None
        ${ids}=    Create List
        Set Global Variable    ${SEEDED_JOB_IDS}    ${ids}
    END
    Append To List    ${ids}    ${id}

Ensure Job Under Test
    [Documentation]    Creates the job the whole jobscheduler/job suite exercises, so
    ...    that any single file can also be run on its own.
    ${id}=    Get Variable Value    ${JOB_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}    ${payload}=    Create Job
    Set Global Variable    ${JOB_ID}    ${id}
    Set Global Variable    ${JOB_ETAG}    ${etag}
    Set Global Variable    ${JOB_KEY}    ${payload}[job_key]

Ensure Seeded Jobs
    [Documentation]    Registers ${count} jobs so the search suite has more than one
    ...    page to work with.
    [Arguments]    ${count}=12
    ${done}=    Get Variable Value    ${JOBS_SEEDED}    ${EMPTY}
    IF    $done    RETURN
    Ensure Job Under Test
    FOR    ${i}    IN RANGE    ${count}
        Create Job
    END
    Set Global Variable    ${JOBS_SEEDED}    ${TRUE}

Ensure Execution History Exists
    [Documentation]    Answers an execution id from the seeded history, skipping the
    ...    test when the database has none.
    ...
    ...    Executions are written by the engine, never by the API, so a read-only suite
    ...    cannot create its own fixture. Skipping is honest where inventing one would
    ...    not be: it says the contract was not exercised rather than implying it passed.
    ${id}=    Get Variable Value    ${EXECUTION_ID}    ${EMPTY}
    IF    $id
        RETURN    ${id}
    END
    ${resp}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    len($items) == 0
        Skip    No execution history in this database; run the scheduler first.
    END
    Set Global Variable    ${EXECUTION_ID}    ${items}[0][id]
    RETURN    ${items}[0][id]

Delete Jobscheduler Seed Data
    [Documentation]    Best effort: the suite must not fail in teardown, and a job
    ...    already removed by 08_delete.robot is the expected case rather than an error.
    ${ids}=    Get Variable Value    ${SEEDED_JOB_IDS}    ${None}
    IF    $ids is not None
        FOR    ${id}    IN    @{ids}
            DELETE On Session    api    ${JOB_API}/${id}    expected_status=any
        END
    END
    # The module-wide delete is the backstop for anything the list missed - a job whose
    # create succeeded but whose id never reached the teardown list.
    DELETE On Session    api    ${JOB_API}
    ...    params=${{ {'module_name': $ROBOT_MODULE} }}    expected_status=any
    Set Global Variable    ${SEEDED_JOB_IDS}    ${None}
    Set Global Variable    ${JOB_ID}    ${EMPTY}
    Set Global Variable    ${JOBS_SEEDED}    ${EMPTY}
