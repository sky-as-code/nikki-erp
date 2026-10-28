*** Settings ***
Documentation     Audit Log helpers: the history of one record, read through
...               GET /v1/auditlog/actions/history, and the audited UoM category the suites
...               mutate. All keywords expect an authorized session with alias "api" unless
...               they take a session alias.
Library           Collections
Resource          ${CURDIR}/essential.resource


*** Variables ***
${AUDIT_HISTORY_API}        /v1/auditlog/actions/history
${AUDIT_ACTION_API}         /v1/auditlog/auditlog_action
${AUDITLOG_SCHEMA_DIR}      ${CURDIR}/schemas/auditlog
${AUDIT_RESOURCE_CODE}      essential_uomcat


*** Keywords ***
Ensure Audited Category
    [Documentation]    A UoM category of its own, so the suite's history starts empty and
    ...    is shared with no other suite. Lazy so any suite file runs standalone.
    ${id}=    Get Variable Value    ${AUDITED_CAT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Org Under Test
    ${name}=    Unique Display Name    Robot Audited Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${AUDITED_CAT_ID}    ${id}
    Set Global Variable    ${AUDITED_CAT_ETAG}    ${etag}
    Set Global Variable    ${AUDITED_CAT_NAME}    ${name}
    Set Global Variable    ${AUDITED_CAT_REQUEST_ID}    ${resp.headers}[X-Request-Id]

Get History Response
    [Documentation]    Raw history call, for the tests that expect a refusal.
    [Arguments]    ${params}    ${session}=api
    ${resp}=    GET On Session    ${session}    ${AUDIT_HISTORY_API}    params=${params}    expected_status=any
    RETURN    ${resp}

Get History
    [Documentation]    Reads one record's history, newest first, and validates the envelope
    ...    against the AUDIT-007 contract. Returns the response body.
    [Arguments]    ${resource_id}    ${resource_code}=${AUDIT_RESOURCE_CODE}    ${session}=api
    ...    ${org_id}=${UOM_ORG_ID}    ${page}=0    ${size}=20
    ${params}=    Create Dictionary    org_id=${org_id}    resource_code=${resource_code}
    ...    resource_id=${resource_id}    page=${page}    size=${size}
    ${resp}=    Get History Response    ${params}    ${session}
    Response Should Match Schema    ${resp}    ${AUDITLOG_SCHEMA_DIR}/history.json    200
    RETURN    ${resp.json()}

Latest Entry
    [Documentation]    The newest entry of a record's history.
    [Arguments]    ${resource_id}
    ${body}=    Get History    ${resource_id}
    Should Not Be Empty    ${body}[items]    msg=The record has no history
    RETURN    ${body}[items][0]

Entry Should Be
    [Documentation]    Asserts the action and the exact set of changed fields of one entry.
    [Arguments]    ${entry}    ${action}    @{fields}
    Should Be Equal    ${entry}[action]    ${action}
    ${changed}=    Evaluate    sorted(change['field'] for change in $entry['changes'])
    ${expected}=    Evaluate    sorted($fields)
    Lists Should Be Equal    ${changed}    ${expected}
    ...    msg=Entry ${entry}[action] should change exactly ${expected}, got ${changed}

Change Of
    [Documentation]    The change of one field in an entry.
    [Arguments]    ${entry}    ${field}
    FOR    ${change}    IN    @{entry}[changes]
        IF    $change['field'] == $field    RETURN    ${change}
    END
    Fail    Entry ${entry}[action] has no change of ${field}

History Count
    [Arguments]    ${resource_id}
    ${body}=    Get History    ${resource_id}
    RETURN    ${body}[total]

Update Audited Category Name
    [Documentation]    Patches the name, rotating the saved etag.
    [Arguments]    ${name}
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${AUDITED_CAT_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $AUDITED_CAT_ETAG, 'org_id': $UOM_ORG_ID} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${AUDITED_CAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${AUDITED_CAT_ETAG}    ${etag}
    RETURN    ${resp}

Set Audited Category Archived
    [Arguments]    ${archived}
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${AUDITED_CAT_ID}/archived
    ...    json=${{ {'org_id': $UOM_ORG_ID, 'etag': $AUDITED_CAT_ETAG, 'is_archived': $archived} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${AUDITED_CAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${AUDITED_CAT_ETAG}    ${etag}

Delete Audit Fixtures
    [Documentation]    Best-effort removal of what the suites created and did not delete.
    FOR    ${name}    IN    AUDITED_CAT_ID    BULK_CAT_A_ID    BULK_CAT_B_ID
        ${id}=    Get Variable Value    ${${name}}    ${EMPTY}
        IF    $id
            DELETE On Session    api    ${UOMCAT_API}/${id}
            ...    params=${{ {'org_id': $UOM_ORG_ID} }}    expected_status=any
            Set Global Variable    ${${name}}    ${EMPTY}
        END
    END
