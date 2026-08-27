*** Settings ***
Documentation     Job Scheduler/Job - Exists.
Resource          resources/jobscheduler.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Job Under Test
Suite Teardown    Delete Jobscheduler Seed Data
Test Tags         jobscheduler    job    exists


*** Test Cases ***
Exists With A Known Id Succeeds
    ${resp}=    POST On Session    api    ${JOB_API}/exists
    ...    json=${{ {'ids': [$JOB_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With An Unknown Id Reports It Missing
    ${resp}=    POST On Session    api    ${JOB_API}/exists
    ...    json=${{ {'ids': [$NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=0    not_existing=1

Exists Separates Known From Unknown
    ${ids}=    Not Found Id List    3
    ${mixed}=    Evaluate    [$JOB_ID] + $ids
    ${resp}=    POST On Session    api    ${JOB_API}/exists    json=${{ {'ids': $mixed} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=3

Exists With A Malformed Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${JOB_API}/exists
    ...    json=${{ {'ids': ['not-a-ulid']} }}    expected_status=any
    Response Status Should Be    ${resp}    400
