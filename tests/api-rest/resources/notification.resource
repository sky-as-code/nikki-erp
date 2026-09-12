*** Settings ***
Documentation     Notification helpers shared by the notification suites: the organization
...               the fixtures belong to, and the inbox reads every suite performs. All
...               keywords expect an authorized session with alias "api".
...
...               THERE IS NO CREATE HELPER, and that is the contract rather than an
...               omission. A notification is raised by the module that has something to
...               say, through SendNotification or the command bus; neither has an HTTP
...               route, and 1008002_notification_iam.sql seeds no create action on any of
...               the three resources. A suite that could POST a notification would be
...               testing an endpoint that by design does not exist.
...
...               The suites therefore read an inbox they did not fill. Where the database
...               has no notifications, tests skip rather than invent a fixture: skipping
...               says the contract was not exercised, where a hand-written row would imply
...               it passed.
...
...               There is no archive helper either. None of the three schemas is
...               archivable (BR 6, BR 8) and the seeds grant no set_archived action.
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${NOTIFICATION_API}         /v1/notification/notification_notification
${RECIPIENT_API}            /v1/notification/notification_recipient
${DELIVERY_API}             /v1/notification/notification_delivery

${INBOX_API}                /v1/notification/notifications
${UNREAD_COUNT_API}         /v1/notification/notifications/unread-count
${MARK_READ_API}            /v1/notification/notifications/mark-read
${STREAM_API}               /v1/notification/notifications/stream

${NOTIFICATION_SCHEMA_DIR}  ${CURDIR}/schemas/notification

${NOTIF_ORG_ID}             ${EMPTY}


*** Keywords ***
Ensure Notification Org
    [Documentation]    Every notification resource is org-scoped, and an org-scoped read
    ...    without org_id filters on the empty string: it returns nothing and looks like an
    ...    empty inbox rather than like the missing precondition it is. ORG_ID from the
    ...    environment wins when set; otherwise the first organization the API returns is
    ...    used, which keeps the suite app-agnostic instead of hardcoding a seeded ULID.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${NOTIF_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${NOTIF_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to scope the Notification suites; set ORG_ID in the environment
    Set Global Variable    ${NOTIF_ORG_ID}    ${items}[0][id]

Get Inbox
    [Documentation]    One page of the calling user's own inbox.
    [Arguments]    &{query}
    Ensure Notification Org
    ${params}=    Create Dictionary    org_id=${NOTIF_ORG_ID}
    Set To Dictionary    ${params}    &{query}
    ${resp}=    GET On Session    api    ${INBOX_API}    params=${params}    expected_status=any
    RETURN    ${resp}

Get Unread Count
    Ensure Notification Org
    ${resp}=    GET On Session    api    ${UNREAD_COUNT_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID} }}    expected_status=any
    RETURN    ${resp}

Mark Notifications Read
    [Documentation]    Marks the given ids read for the calling user. Ids belonging to
    ...    somebody else are silently not updated rather than refused (BR 21), so this
    ...    keyword never treats a foreign id as an error.
    [Arguments]    @{notification_ids}
    Ensure Notification Org
    ${resp}=    POST On Session    api    ${MARK_READ_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID} }}
    ...    json=${{ {'notification_ids': list($notification_ids)} }}
    ...    expected_status=any
    RETURN    ${resp}

Ensure Notification In Inbox
    [Documentation]    Answers one notification id from the calling user's inbox, skipping
    ...    the test when there is none.
    ...
    ...    Notifications are written by the module that raises them, never by the API, so a
    ...    read-only suite cannot create its own fixture. Skipping is honest where inventing
    ...    one would not be: it says the contract was not exercised rather than implying it
    ...    passed.
    ${id}=    Get Variable Value    ${NOTIFICATION_ID}    ${EMPTY}
    IF    $id
        RETURN    ${id}
    END
    ${resp}=    Get Inbox    size=1
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    len($items) == 0
        Skip    No notifications in this inbox; raise one through a source module first.
    END
    Set Global Variable    ${NOTIFICATION_ID}    ${items}[0][notification_id]
    RETURN    ${items}[0][notification_id]

Ensure Unread Notification In Inbox
    [Documentation]    Answers one UNREAD notification id, skipping when every notification
    ...    has already been read. Separate from the keyword above because the mark-read
    ...    tests need a notification whose read_at is still null, and reusing a read one
    ...    would assert nothing.
    ${resp}=    Get Inbox    is_read=${FALSE}    size=1
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    len($items) == 0
        Skip    No unread notifications in this inbox; raise one through a source module first.
    END
    RETURN    ${items}[0][notification_id]
