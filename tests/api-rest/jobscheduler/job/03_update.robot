*** Settings ***
Documentation     Job Scheduler/Job - Update. Uses the job under test created by
...               02_create.robot, or creates one when this file runs on its own.
Resource          resources/jobscheduler.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Job Under Test
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    update


*** Test Cases ***
Update Name Succeeds
    ${name}=    Unique Display Name    Robot Renamed Job
    ${resp}=    PUT On Session    api    ${JOB_API}/${JOB_ID}
    ...    json=${{ {'name': $name, 'etag': $JOB_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${JOB_ETAG}
    Set Global Variable    ${JOB_ETAG}    ${etag}

Update Cron Recomputes The Next Run Instant
    [Documentation]    Editing the schedule must move next_run_at, or the engine would keep
    ...    waking at the instant the old expression implied and run the job on a schedule
    ...    nobody asked for.
    ${before}=    GET On Session    api    ${JOB_API}/${JOB_ID}
    ${old}=    Set Variable    ${before.json()}[item][next_run_at]

    ${resp}=    PUT On Session    api    ${JOB_API}/${JOB_ID}
    ...    json=${{ {'cron_expression': '3 4 * * *', 'etag': $JOB_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    Set Global Variable    ${JOB_ETAG}    ${etag}

    ${after}=    GET On Session    api    ${JOB_API}/${JOB_ID}
    ${new}=    Set Variable    ${after.json()}[item][next_run_at]
    Should Not Be Equal    ${new}    ${old}
    ...    msg=next_run_at must be recomputed when the cron expression changes
    Should Contain    ${new}    T04:03:00    msg=The new instant must follow the new expression

Disabling A Job Clears Its Next Run Instant
    [Documentation]    A null next_run_at is how a job leaves the engine's index. Leaving a
    ...    value behind would have the timer wake for work that will never be created.
    ${id}    ${etag}    ${payload}=    Create Job
    ${resp}=    PUT On Session    api    ${JOB_API}/${id}
    ...    json=${{ {'is_enabled': False, 'etag': $etag} }}
    Response Should Be Update Success    ${resp}    count=1

    ${after}=    GET On Session    api    ${JOB_API}/${id}
    ${item}=    Item Should Match Schema    ${after}    ${JOBSCHEDULER_SCHEMA_DIR}/job.json    200
    ${next_run_at}=    Evaluate    $item.get('next_run_at')
    Should Be Equal    ${next_run_at}    ${None}
    ...    msg=A disabled job must hold no next_run_at

Update With Stale Etag Fails
    [Tags]    negative
    ${resp}=    PUT On Session    api    ${JOB_API}/${JOB_ID}
    ...    json=${{ {'name': 'Whatever', 'etag': 'stale-etag-value'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Cron Fails
    [Tags]    negative
    ${resp}=    PUT On Session    api    ${JOB_API}/${JOB_ID}
    ...    json=${{ {'cron_expression': '@hourly', 'etag': $JOB_ETAG} }}    expected_status=any
    Response Status Should Be    ${resp}    400

Update Inverting The Effective Period Fails
    [Tags]    negative
    [Documentation]    A partial update sending only effective_until must still be checked
    ...    against the stored effective_from. Validating the incoming fields alone would let
    ...    an edit that inverts the period pass, because from its own point of view there is
    ...    no period at all.
    ${id}    ${etag}    ${payload}=    Create Job
    ...    &{{ {'effective_from': '2026-08-24T10:00:00Z'} }}

    ${resp}=    PUT On Session    api    ${JOB_API}/${id}
    ...    json=${{ {'effective_until': '2026-08-24T09:00:00Z', 'etag': $etag} }}
    ...    expected_status=any
    Response Status Should Be    ${resp}    400

Update Of Unknown Job Fails
    [Tags]    negative
    ${resp}=    PUT On Session    api    ${JOB_API}/${NOT_FOUND_ID}
    ...    json=${{ {'name': 'Ghost', 'etag': 'any'} }}    expected_status=any
    Response Status Should Be    ${resp}    400
