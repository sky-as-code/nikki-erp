*** Settings ***
Documentation     Job Scheduler/Job - Search.
Resource          resources/jobscheduler.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seeded Jobs    12
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    search


*** Variables ***
${JOB_SCHEMA}    ${JOBSCHEDULER_SCHEMA_DIR}/job.json


*** Test Cases ***
Search Without Criteria Returns The Default Page
    ${resp}=    GET On Session    api    ${JOB_API}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50    page=0

Search With Paging Returns The Requested Page
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'page': 1, 'size': 5} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=5    page=1

Search Beyond The Last Page Returns Nothing
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'page': 999, 'size': 5} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=5    page=999    item_count=0

Search By Module Name Returns Only That Module
    ${graph}=    Set Variable    {"if": ["module_name", "=", "${ROBOT_MODULE}"]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph, 'size': 50} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[module_name]    ${ROBOT_MODULE}
    END

Search By Job Key Finds The Job Under Test
    ${graph}=    Set Variable    {"if": ["job_key", "=", "${JOB_KEY}"]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50    item_count=1
    Search Results Should Contain Id    ${resp}    ${JOB_ID}

Search By Job Type Returns Only Technical Jobs
    ${graph}=    Set Variable    {"if": ["job_type", "=", "technical"]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph, 'size': 50} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[job_type]    technical
    END

Search With A Contains Filter Narrows The Result
    ${graph}=    Set Variable    {"if": ["job_key", "*", "robot_job"]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph, 'size': 50} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Contain    ${item}[job_key]    robot_job
    END

Search With An And Graph Applies Both Conditions
    ${graph}=    Set Variable
    ...    {"and": [{"if": ["module_name", "=", "${ROBOT_MODULE}"]}, {"if": ["job_type", "=", "technical"]}]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph, 'size': 50} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[module_name]    ${ROBOT_MODULE}
        Should Be Equal    ${item}[job_type]    technical
    END

Search Can Order By Job Key
    ${graph}=    Set Variable
    ...    {"and": [{"if": ["module_name", "=", "${ROBOT_MODULE}"]}], "order": [["job_key", "asc"]]}
    ${resp}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph, 'size': 50} }}
    Response Should Be Search Success    ${resp}    ${JOB_SCHEMA}    size=50
    ${keys}=    Evaluate    [i['job_key'] for i in $resp.json()['items']]
    Should Be Equal    ${keys}    ${{ sorted($keys) }}
    ...    msg=Results must come back in the order the graph asked for

Search With A Nonexistent Field Fails
    [Tags]    negative
    ${graph}=    Set Variable    {"if": ["bla_bla_field", "=", "x"]}
    ${resp}=    GET On Session    api    ${JOB_API}
    ...    params=${{ {'graph': $graph} }}    expected_status=any
    Response Status Should Be    ${resp}    400
