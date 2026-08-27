*** Settings ***
Documentation     Job Scheduler/Job - Delete, and the cleanup for the whole resource suite.
...               Runs last so the earlier files still have their job under test.
...
...               It also carries the assertions that would have lived in 06_archive.robot:
...               a scheduled job is deleted permanently and never archived, and that
...               absence is asserted here rather than left as a missing file.
Resource          resources/jobscheduler.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    delete


*** Test Cases ***
Delete Job Succeeds
    ${id}    ${etag}    ${payload}=    Create Job
    ${resp}=    DELETE On Session    api    ${JOB_API}/${id}
    Response Should Be Delete Success    ${resp}    count=1

    ${after}=    GET On Session    api    ${JOB_API}/${id}    expected_status=any
    Response Should Be Not Found Error    ${after}

Deleting A Job Keeps Its Execution History
    [Documentation]    Deleting a registration must not destroy the record of what it did.
    ...    The execution's job_id is nulled by the foreign key rather than by application
    ...    code, so this holds even for a code path that forgets about history.
    ${id}    ${etag}    ${payload}=    Create Job
    ${before}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'size': 1} }}
    Response Status Should Be    ${before}    200
    ${total_before}=    Set Variable    ${before.json()}[total]

    DELETE On Session    api    ${JOB_API}/${id}

    ${after}=    GET On Session    api    ${EXECUTION_API}    params=${{ {'size': 1} }}
    Response Status Should Be    ${after}    200
    Should Be Equal As Integers    ${after.json()}[total]    ${total_before}
    ...    msg=Deleting a job must not remove any execution row

Delete Unknown Job Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${JOB_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

# --- Module-wide delete ---

Delete By Module Removes Every Job Of That Module
    ${id}    ${etag}    ${payload}=    Create Job
    ${resp}=    DELETE On Session    api    ${JOB_API}
    ...    params=${{ {'module_name': $ROBOT_MODULE} }}
    Response Status Should Be    ${resp}    200
    Should Be True    ${resp.json()}[affected_count] >= 1

    ${graph}=    Set Variable    {"if": ["module_name", "=", "${ROBOT_MODULE}"]}
    ${after}=    GET On Session    api    ${JOB_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${after}    200
    Length Should Be    ${after.json()}[items]    0

Delete By Module Is Idempotent
    [Documentation]    Deleting nothing is a success, not a 404. A module uninstalling
    ...    itself should not have to know whether it ever registered anything, and making
    ...    the caller distinguish the two cases would only produce callers that ignore it.
    DELETE On Session    api    ${JOB_API}    params=${{ {'module_name': $ROBOT_MODULE} }}
    ${resp}=    DELETE On Session    api    ${JOB_API}
    ...    params=${{ {'module_name': $ROBOT_MODULE} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[affected_count]    0

Delete By Module Without A Module Name Fails
    [Tags]    negative
    [Documentation]    The empty name is the one that matters. It is one keystroke from a
    ...    real one, and treating it as "all modules" would make a typo delete every
    ...    registration in the system.
    ${resp}=    DELETE On Session    api    ${JOB_API}    expected_status=any
    Response Status Should Be    ${resp}    400

Delete By Module With An Empty Module Name Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${JOB_API}
    ...    params=${{ {'module_name': ''} }}    expected_status=any
    Response Status Should Be    ${resp}    400

Delete By Module Leaves Other Modules Alone
    [Documentation]    The match is exact, with no wildcard, so a module name that merely
    ...    shares a prefix with another must survive.
    ${keep}    ${etag}    ${payload}=    Create Job
    ...    &{{ {'module_name': 'robotjobschedulerother'} }}
    DELETE On Session    api    ${JOB_API}    params=${{ {'module_name': $ROBOT_MODULE} }}

    ${resp}=    GET On Session    api    ${JOB_API}/${keep}
    Response Status Should Be    ${resp}    200
    [Teardown]    DELETE On Session    api    ${JOB_API}
    ...    params=${{ {'module_name': 'robotjobschedulerother'} }}    expected_status=any

# --- A job is not archivable ---

Job Has No Archive Endpoint
    [Tags]    negative
    [Documentation]    Stands in for 06_archive.robot. A scheduled job is removed
    ...    permanently or not at all: 0002002_jobscheduler_iam.sql seeds no set_archived
    ...    action and the schema declines the archivable mixin, so the route must not exist.
    ...    Asserting that here keeps the decision visible as a test rather than as a gap in
    ...    the file numbering.
    ${id}    ${etag}    ${payload}=    Create Job
    ${resp}=    POST On Session    api    ${JOB_API}/${id}/archived
    ...    json=${{ {'is_archived': True, 'etag': $etag} }}    expected_status=any
    Should Contain    ${{ [404, 405] }}    ${resp.status_code}
    ...    msg=Expected no archive route, got ${resp.status_code}: ${resp.text}
