*** Settings ***
Documentation     Notification Inbox - the listing, its filters and its ordering (BR 18).
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    inbox


*** Variables ***
${INBOX_ITEM_SCHEMA}    ${NOTIFICATION_SCHEMA_DIR}/inbox_item.json


*** Test Cases ***
Get The Inbox
    ${resp}=    Get Inbox
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    items

Every Inbox Item Matches The Item Contract
    [Documentation]    The inbox item is not the notification row: it joins the
    ...    notification to the caller's own recipient row, and carries the calculated
    ...    is_read alongside the stream sequence the client resumes from.
    ${resp}=    Get Inbox    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Validate Json Schema    ${item}    ${INBOX_ITEM_SCHEMA}
    END

The Inbox Is Ordered Newest First
    [Documentation]    BR 18. stream_seq descending, and NOT created_at: two notifications
    ...    raised in the same instant still have a defined order, which a timestamp cannot
    ...    give (BR 27).
    ${resp}=    Get Inbox    limit=20
    Response Status Should Be    ${resp}    200
    ${seqs}=    Evaluate    [i['stream_seq'] for i in $resp.json()['items']]
    Should Be Equal    ${seqs}    ${{ sorted($seqs, reverse=True) }}

Is Read False Returns Only Unread Notifications
    ${resp}=    Get Inbox    is_read=${FALSE}    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[is_read]    ${FALSE}
        # read_at is omitted rather than sent as null when the notification is unread, so
        # its ABSENCE is what "unread" looks like on the wire.
        ${has_read_at}=    Evaluate    $item.get('read_at') is not None
        Should Be Equal    ${has_read_at}    ${FALSE}
    END

Is Read True Returns Only Read Notifications
    ${resp}=    Get Inbox    is_read=${TRUE}    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[is_read]    ${TRUE}
        ${has_read_at}=    Evaluate    $item.get('read_at') is not None
        Should Be Equal    ${has_read_at}    ${TRUE}
    END

Is Read Is Calculated From Read At
    [Documentation]    BR 7. is_read is never stored, so the two can never disagree. This
    ...    is what the test is really asserting: not that both are present, but that one is
    ...    derived from the other.
    ${resp}=    Get Inbox    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${expected}=    Evaluate    $item.get('read_at') is not None
        Should Be Equal    ${item}[is_read]    ${expected}
    END

Filtering By Severity Returns Only That Severity
    ${resp}=    Get Inbox    severity=warning    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[severity]    warning
    END

Filtering By Source Module Returns Only That Module
    ${resp}=    Get Inbox    source_module=sales    limit=20
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[source_module]    sales
    END

The Limit Bounds The Page
    ${resp}=    Get Inbox    limit=2
    Response Status Should Be    ${resp}    200
    ${count}=    Get Length    ${resp.json()}[items]
    Should Be True    ${count} <= 2

The Cursor Pages Strictly Backwards
    [Documentation]    The cursor is a position in the stream, not an offset: paging by
    ...    offset shifts under a reader whenever a notification arrives at the top, which
    ...    is exactly when someone is reading.
    ${first}=    Get Inbox    limit=1
    Response Status Should Be    ${first}    200
    IF    len($first.json()['items']) == 0
        Skip    No notifications in this inbox; raise one through a source module first.
    END
    ${cursor}=    Set Variable    ${first.json()}[items][0][stream_seq]

    ${second}=    Get Inbox    limit=5    cursor=${cursor}
    Response Status Should Be    ${second}    200
    FOR    ${item}    IN    @{second.json()}[items]
        Should Be True    ${item}[stream_seq] < ${cursor}
        ...    msg=A cursor page must contain only notifications older than the cursor
    END

An Unauthenticated Request Is Refused
    [Documentation]    AC-FS-02, and the same rule for the REST reads: the inbox is
    ...    somebody's private correspondence, so an anonymous caller must get nothing.
    Create Anonymous API Session    alias=notif_anon
    ${resp}=    GET On Session    notif_anon    ${INBOX_API}    expected_status=any
    Should Be True    ${resp.status_code} in (401, 403)
