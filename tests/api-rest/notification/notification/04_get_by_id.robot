*** Settings ***
Documentation     Notification - Get by id, and the isolation that read must enforce.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    get_by_id


*** Variables ***
${NOTIFICATION_SCHEMA}    ${NOTIFICATION_SCHEMA_DIR}/notification.json


*** Test Cases ***
Get A Notification By Id
    ${id}=    Ensure Notification In Inbox
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}/${id}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${NOTIFICATION_SCHEMA}
    Should Be Equal    ${item}[id]    ${id}

Get An Unknown Notification Answers Not Found
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}/${NOT_FOUND_ID}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} in (400, 404)

Reading Another Organization's Notification Is Refused
    [Documentation]    AC14. A notification belongs to one organization, and naming another
    ...    one must not reach it. The answer may be a refusal or a not-found — either is
    ...    acceptable, and which one matters far less than that the row is never returned.
    ${id}=    Ensure Notification In Inbox
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}/${id}
    ...    params=${{ {'org_id': $NOT_FOUND_ID} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A notification was returned for an organization the caller does not belong to
