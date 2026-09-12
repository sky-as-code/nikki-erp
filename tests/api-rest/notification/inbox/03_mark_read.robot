*** Settings ***
Documentation     Notification Inbox - mark read, single and bulk (BR 20, BR 21).
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    inbox    mark_read


*** Test Cases ***
Mark One Notification Read
    ${id}=    Ensure Unread Notification In Inbox
    ${resp}=    Mark Notifications Read    ${id}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[updated_count]    1

    ${after}=    Get Inbox    limit=100
    Response Status Should Be    ${after}    200
    ${item}=    Find Inbox Item    ${after}    ${id}
    Should Be Equal    ${item}[is_read]    ${TRUE}
    ${has_read_at}=    Evaluate    $item.get('read_at') is not None
    Should Be Equal    ${has_read_at}    ${TRUE}

Marking Read Twice Is Idempotent And Keeps The First Timestamp
    [Documentation]    AC07, BR 20.4-20.6. The second call must change nothing, and must
    ...    NOT overwrite read_at: the first read is the one that happened, and rewriting the
    ...    timestamp would quietly make "when did I see this" unanswerable.
    ${id}=    Ensure Unread Notification In Inbox

    ${first}=    Mark Notifications Read    ${id}
    Response Status Should Be    ${first}    200
    Should Be Equal As Integers    ${first.json()}[updated_count]    1

    ${inbox}=    Get Inbox    limit=100
    ${first_item}=    Find Inbox Item    ${inbox}    ${id}
    ${original}=    Set Variable    ${first_item}[read_at]

    ${second}=    Mark Notifications Read    ${id}
    Response Status Should Be    ${second}    200
    Should Be Equal As Integers    ${second.json()}[updated_count]    0
    Should Be Equal As Integers    ${second.json()}[already_read_count]    1

    ${after}=    Get Inbox    limit=100
    ${second_item}=    Find Inbox Item    ${after}    ${id}
    Should Be Equal    ${second_item}[read_at]    ${original}
    ...    msg=A second mark-read must not overwrite the original read_at

Marking Read Lowers The Unread Count By One
    ${id}=    Ensure Unread Notification In Inbox
    ${before}=    Get Unread Count
    Response Status Should Be    ${before}    200
    ${was}=    Set Variable    ${before.json()}[count]

    ${resp}=    Mark Notifications Read    ${id}
    Response Status Should Be    ${resp}    200

    ${after}=    Get Unread Count
    Response Status Should Be    ${after}    200
    Should Be Equal As Integers    ${after.json()}[count]    ${{ $was - 1 }}

A Notification Id That Is Not The Caller's Is Not Updated
    [Documentation]    BR 21. A valid id belonging to somebody else is silently not
    ...    updated rather than refusing the whole batch — one stray id must not lose every
    ...    real read in it. requested_count counts what was asked for; updated_count counts
    ...    what was the caller's to change.
    ${resp}=    Mark Notifications Read    ${NOT_FOUND_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[requested_count]    1
    Should Be Equal As Integers    ${resp.json()}[updated_count]    0
    Should Be Equal As Integers    ${resp.json()}[already_read_count]    0

A Mixed Batch Updates Only The Caller's Own
    ${id}=    Ensure Unread Notification In Inbox
    ${resp}=    Mark Notifications Read    ${id}    ${NOT_FOUND_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[requested_count]    2
    Should Be Equal As Integers    ${resp.json()}[updated_count]    1

An Empty Id List Is Rejected
    [Documentation]    Asking to mark nothing read is a mistake rather than an
    ...    instruction, and answering 200 would let a broken client believe it had done
    ...    something.
    ${resp}=    Mark Notifications Read
    Should Be Equal As Integers    ${resp.status_code}    400

An Unauthenticated Request Is Refused
    Create Anonymous API Session    alias=notif_anon_mark
    ${resp}=    POST On Session    notif_anon_mark    ${MARK_READ_API}
    ...    json=${{ {'notification_ids': ['x']} }}    expected_status=any
    Should Be True    ${resp.status_code} in (401, 403)


*** Keywords ***
Find Inbox Item
    [Documentation]    The one inbox item carrying this notification id.
    ...
    ...    The id is passed as an argument rather than read through Evaluate's $name syntax:
    ...    that syntax resolves against Robot's variable scope, which a keyword-returned
    ...    local is not reliably part of.
    [Arguments]    ${resp}    ${notification_id}
    ${items}=    Set Variable    ${resp.json()}[items]
    ${item}=    Evaluate
    ...    next((i for i in $items if i['notification_id'] == '${notification_id}'), None)
    Should Not Be Equal    ${item}    ${None}
    ...    msg=Notification ${notification_id} is not in the inbox page
    RETURN    ${item}
