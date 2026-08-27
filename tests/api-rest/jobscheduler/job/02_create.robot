*** Settings ***
Documentation     Job Scheduler/Job - Create. The first test saves the "job under test"
...               (${JOB_ID}/${JOB_ETAG}/${JOB_KEY} globals) consumed by the later suites
...               and deleted last by 08_delete.robot.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${payload}=    New Job Payload
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Remember Seeded Job    ${id}
    Set Global Variable    ${JOB_ID}    ${id}
    Set Global Variable    ${JOB_ETAG}    ${etag}
    Set Global Variable    ${JOB_KEY}    ${payload}[job_key]

Create Computes The Next Run Instant
    [Documentation]    next_run_at is what the engine's index finds a due job by. A job
    ...    registered without one is invisible to the scheduler: it would sit enabled and
    ...    correct-looking and never fire.
    ${id}    ${etag}    ${payload}=    Create Job    cron=*/5 * * * *
    ${resp}=    GET On Session    api    ${JOB_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${JOBSCHEDULER_SCHEMA_DIR}/job.json    200
    Should Not Be Empty    ${item}[next_run_at]
    Should End With    ${item}[next_run_at]    Z    msg=next_run_at must be serialized as UTC

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${JOB_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}
    ...    name    job_type    module_name    job_key    action_type    action_config    cron_expression

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${JOB_API}
    ...    data={ "name": "",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

# --- Cron expression (the module's own rule, not the schema's) ---

Create With Six Field Cron Fails
    [Tags]    negative
    [Documentation]    Six fields means a seconds column. The scheduler's minimum resolution
    ...    is one minute, so accepting it would silently run the job on a schedule the
    ...    caller did not ask for.
    ${payload}=    New Job Payload    cron=0 0 12 * * *
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    cron_expression

Create With Quartz Placeholder Cron Fails
    [Tags]    negative
    [Documentation]    "?" is Quartz syntax that people type reflexively. It is not supported,
    ...    and must be refused at registration rather than at the first occurrence.
    ${payload}=    New Job Payload    cron=0 0 * * ?
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    cron_expression

Create With Descriptor Cron Fails
    [Tags]    negative
    ${payload}=    New Job Payload    cron=@daily
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    cron_expression

# --- Effective period ---

Create With Effective Until Before From Fails
    [Tags]    negative
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}
    ...    effective_from=2026-08-24T10:00:00Z    effective_until=2026-08-24T09:00:00Z
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    effective_until

Create With Equal Effective Bounds Fails
    [Tags]    negative
    [Documentation]    The period is half-open, [from, until), so equal bounds describe a
    ...    window in which nothing can ever fire. A job that is registered and permanently
    ...    silent is the hardest kind of failure to notice, so it is refused.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}
    ...    effective_from=2026-08-24T10:00:00Z    effective_until=2026-08-24T10:00:00Z
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    effective_until

Create With Non Utc Datetime Fails
    [Tags]    negative
    [Documentation]    Every instant the scheduler stores is UTC. An offset-bearing datetime
    ...    is rejected rather than converted, so that what a caller wrote and what the
    ...    scheduler runs cannot differ.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    effective_from=2026-08-24T10:00:00+07:00
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400

# --- Field constraints the schema owns ---

Create With Non Alphanumeric Module Name Fails
    [Tags]    negative
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    module_name=my-module
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    module_name

Create With Retry Interval Below The Floor Fails
    [Tags]    negative
    [Documentation]    Five seconds is the floor, and it is rejected rather than normalized:
    ...    silently raising a caller's 3 to 5 would make the job behave differently from
    ...    what its own row says.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    retry_interval_seconds=${3}
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Should Be Invalid Number Range Error    ${resp}    retry_interval_seconds

Create With Unknown Action Type Fails
    [Tags]    negative
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    action_type=smoke_signal
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    action_type

Create With Unsupported Http Method Fails
    [Tags]    negative
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}
    ...    action_config=${{ {'method': 'TRACE', 'url': 'https://api.nikkierp.com/v1/health'} }}
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400

Create With Non Http Url Fails
    [Tags]    negative
    [Documentation]    A file:// action would turn an authenticated job registration into
    ...    local file access on the scheduler host, so the scheme is restricted.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}
    ...    action_config=${{ {'method': 'GET', 'url': 'file:///etc/passwd'} }}
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400

Create With Unregistered Command Fails
    [Tags]    negative
    [Documentation]    A mistyped command name would otherwise become a job that runs on
    ...    schedule forever, publishes into the void, and reports success every time. There
    ...    is no later point at which the mistake surfaces on its own.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    action_type=command_bus
    ...    action_config=${{ {'command_name': 'inventory_maintenance.definitelyNotRegistered'} }}
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400

Create With User Job Type Fails
    [Tags]    negative
    [Documentation]    The column admits "user" because it will hold it once user-managed
    ...    scheduling exists. Nothing can run such a job today, so accepting one now would
    ...    register a job that is silent forever.
    ${payload}=    New Job Payload
    Set To Dictionary    ${payload}    job_type=user
    ${resp}=    POST On Session    api    ${JOB_API}    json=${payload}    expected_status=any
    Response Status Should Be    ${resp}    400
    Error Should Target Field    ${resp}    job_type

# --- Idempotent registration ---

Repeat Registration Returns The Existing Job
    [Documentation]    A module registers its jobs on every boot, so the second boot must not
    ...    be an error. The repeat answers 200 rather than 201, and the same id, so a caller
    ...    can tell a fresh registration from a no-op.
    ${payload}=    New Job Payload
    ${first}=    POST On Session    api    ${JOB_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${first}
    Remember Seeded Job    ${id}

    ${second}=    POST On Session    api    ${JOB_API}    json=${payload}
    Response Status Should Be    ${second}    200
    Should Be Equal    ${second.json()}[id]    ${id}
    ...    msg=A repeat registration must answer the job that already exists

Repeat Registration Creates No Second Row
    ${payload}=    New Job Payload
    ${first}=    POST On Session    api    ${JOB_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${first}
    Remember Seeded Job    ${id}
    POST On Session    api    ${JOB_API}    json=${payload}

    ${graph}=    Set Variable    {"if": ["job_key", "=", "${payload}[job_key]"]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    1
    ...    msg=(module_name, job_key) must identify exactly one job


*** Keywords ***
Error Should Target Field
    [Documentation]    Asserts a 400 body carries an error naming ${field}.
    ...
    ...    Deliberately looser than the shared Response Should Be * keywords: those pin the
    ...    exact key and message, which is right for the platform-wide contracts they cover.
    ...    These are this module's own rules, and pinning their wording here would make the
    ...    test fail on a reworded message rather than on a behaviour change.
    [Arguments]    ${resp}    ${field}
    ${fields}=    Evaluate    [e.get('field') for e in $resp.json()]
    Should Contain    ${fields}    ${field}
    ...    msg=Expected an error on '${field}'. Body: ${resp.text}
