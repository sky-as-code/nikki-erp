*** Settings ***
Documentation     Job Scheduler/Execution - Search, both the unscoped listing and one
...               job's history.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Test Tags         jobscheduler    execution    search


*** Variables ***
${EXECUTION_SCHEMA}    ${JOBSCHEDULER_SCHEMA_DIR}/execution.json


*** Test Cases ***
Search Without Criteria Returns The Default Page
    ${resp}=    GET On Session    api    ${EXECUTION_API}
    Response Should Be Search Success    ${resp}    ${EXECUTION_SCHEMA}    size=50    page=0
    ...    item_count=${None}

Search With Paging Returns The Requested Page
    ${resp}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'page': 1, 'size': 5} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    5
    Should Be Equal As Integers    ${resp.json()}[page]    1

Search By Status Returns Only That Status
    ${graph}=    Set Variable    {"if": ["status", "=", "failed"]}
    ${resp}=    GET On Session    api    ${EXECUTION_API}
    ...    params=${{ {'graph': $graph, 'size': 20} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[status]    failed
    END

Search Can Order By Scheduled Instant
    ${graph}=    Set Variable    {"order": [["scheduled_for", "desc"]]}
    ${resp}=    GET On Session    api    ${EXECUTION_API}
    ...    params=${{ {'graph': $graph, 'size': 20} }}
    Response Status Should Be    ${resp}    200
    ${instants}=    Evaluate    [i['scheduled_for'] for i in $resp.json()['items']]
    Should Be Equal    ${instants}    ${{ sorted($instants, reverse=True) }}

Every Execution Instant Is Utc
    [Documentation]    The whole module works in UTC. A serialized instant carrying an
    ...    offset would mean a client and the scheduler disagreed about when a job ran.
    ${resp}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'size': 20} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should End With    ${item}[scheduled_for]    Z
        Should End With    ${item}[available_at]    Z
    END

# --- One job's history ---

Search A Job's Executions Returns Only That Job's
    ${graph}=    Set Variable    {"if": ["job_id", "is_set"]}
    ${found}=    GET On Session    api    ${EXECUTION_API}
    ...    params=${{ {'graph': $graph, 'size': 1} }}
    Response Status Should Be    ${found}    200
    IF    len($found.json()['items']) == 0
        Skip    No execution is attached to a job in this database.
    END
    ${job_id}=    Set Variable    ${found.json()}[items][0][job_id]

    ${resp}=    GET On Session    api    ${JOB_API}/${job_id}/executions
    ...    params=${{ {'size': 20} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[job_id]    ${job_id}
    END

A Job's History Cannot Be Widened By A Client Graph
    [Documentation]    The job predicate is ANDed on top of whatever graph the client sent,
    ...    never merged into it. Scoping that the request it scopes can override is not
    ...    scoping at all.
    ${graph}=    Set Variable    {"if": ["job_id", "is_set"]}
    ${found}=    GET On Session    api    ${EXECUTION_API}
    ...    params=${{ {'graph': $graph, 'size': 1} }}
    Response Status Should Be    ${found}    200
    IF    len($found.json()['items']) == 0
        Skip    No execution is attached to a job in this database.
    END
    ${job_id}=    Set Variable    ${found.json()}[items][0][job_id]

    # A graph that on its own would match every execution in the table.
    ${wide}=    Set Variable    {"if": ["status", "!=", "there_is_no_such_status"]}
    ${resp}=    GET On Session    api    ${JOB_API}/${job_id}/executions
    ...    params=${{ {'graph': $wide, 'size': 50} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[job_id]    ${job_id}
        ...    msg=A client graph must not widen the listing past the job in the path
    END

Search An Unknown Job's Executions Returns Nothing
    ${resp}=    GET On Session    api    ${JOB_API}/${NOT_FOUND_ID}/executions
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    0
