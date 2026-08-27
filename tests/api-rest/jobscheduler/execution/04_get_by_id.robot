*** Settings ***
Documentation     Job Scheduler/Execution - Get by id, including the attempt history the
...               detail response composes onto it.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Test Tags         jobscheduler    execution    get


*** Test Cases ***
Get Execution By Id Succeeds
    ${id}=    Ensure Execution History Exists
    ${resp}=    GET On Session    api    ${EXECUTION_API}/${id}
    Response Status Should Be    ${resp}    200
    Validate Json Schema    ${resp.json()}[item]    ${JOBSCHEDULER_SCHEMA_DIR}/execution.json
    Should Be Equal    ${resp.json()}[item][id]    ${id}

Get Execution Returns Its Attempts
    [Documentation]    The detail response carries the attempts beside the execution rather
    ...    than requiring a second call, because the question it answers - why did this end
    ...    the way it did - lives entirely in them.
    ${id}=    Ensure Execution History Exists
    ${resp}=    GET On Session    api    ${EXECUTION_API}/${id}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    attempts
    ${attempts}=    Set Variable    ${resp.json()}[attempts]
    Should Be True    isinstance($attempts, list)
    ...    msg=attempts must always be a list, empty for an execution that has not run yet

Attempts Come Back In The Order They Ran
    [Documentation]    Ordered by attempt_number rather than by time: a reaped attempt whose
    ...    worker died has no reliable finish time, but two attempts of one execution are
    ...    strictly ordered by their number whatever their clocks said.
    ${found}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'size': 50} }}
    Response Status Should Be    ${found}    200
    ${multi}=    Evaluate    [i for i in $found.json()['items'] if i['attempt_count'] > 1]
    IF    len($multi) == 0
        Skip    No multi-attempt execution in this database.
    END

    ${id}=    Set Variable    ${multi}[0][id]
    ${resp}=    GET On Session    api    ${EXECUTION_API}/${id}
    Response Status Should Be    ${resp}    200
    ${numbers}=    Evaluate    [a['attempt_number'] for a in $resp.json()['attempts']]
    Should Be Equal    ${numbers}    ${{ sorted($numbers) }}

A Failed Execution Names Why It Failed
    [Documentation]    failure_code is the machine-readable reason the retry chain ended.
    ...    Without it a failed execution says only that something went wrong, which is the
    ...    one thing an operator already knows.
    ${graph}=    Set Variable    {"if": ["status", "=", "failed"]}
    ${found}=    GET On Session    api    ${EXECUTION_API}
    ...    params=${{ {'graph': $graph, 'size': 1} }}
    Response Status Should Be    ${found}    200
    IF    len($found.json()['items']) == 0
        Skip    No failed execution in this database.
    END
    ${code}=    Set Variable    ${found.json()}[items][0][failure_code]
    Should Contain    ${{ ['RETRY_WINDOW_EXPIRED', 'MAX_ATTEMPTS_REACHED', 'NON_RETRYABLE'] }}
    ...    ${code}    msg=Unexpected failure_code '${code}'

Get Unknown Execution Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${EXECUTION_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Execution Has No Create Route
    [Tags]    negative
    [Documentation]    History is written by the engine alone. A create route would let a
    ...    caller record a run that never happened.
    ${resp}=    POST On Session    api    ${EXECUTION_API}
    ...    json=${{ {'execution_key': 'forged:key:2026-08-24T10:00:00Z'} }}    expected_status=any
    Should Contain    ${{ [404, 405] }}    ${resp.status_code}
    ...    msg=Expected no create route, got ${resp.status_code}: ${resp.text}

Execution Has No Delete Route
    [Tags]    negative
    ${id}=    Ensure Execution History Exists
    ${resp}=    DELETE On Session    api    ${EXECUTION_API}/${id}    expected_status=any
    Should Contain    ${{ [404, 405] }}    ${resp.status_code}
    ...    msg=Expected no delete route, got ${resp.status_code}: ${resp.text}
