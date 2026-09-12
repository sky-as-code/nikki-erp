*** Settings ***
Documentation     Notification Inbox - the unread count (BR 19).
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    inbox    unread_count


*** Test Cases ***
Get The Unread Count
    ${resp}=    Get Unread Count
    Response Status Should Be    ${resp}    200
    ${count}=    Set Variable    ${resp.json()}[count]
    Should Be True    isinstance($count, int) and $count >= 0

The Count Agrees With The Unread Listing
    [Documentation]    BR 19. The count is a COUNT over the same predicate the listing
    ...    filters on, so the two cannot be allowed to drift: a badge showing three unread
    ...    over a list showing one is worse than no badge at all.
    ...
    ...    Compared against a bounded page, so this asserts agreement only while the unread
    ...    set is small enough to fit — beyond that the listing is paged and the comparison
    ...    would be meaningless rather than failing.
    ${count_resp}=    Get Unread Count
    Response Status Should Be    ${count_resp}    200
    ${count}=    Set Variable    ${count_resp.json()}[count]

    ${list_resp}=    Get Inbox    is_read=${FALSE}    limit=100
    Response Status Should Be    ${list_resp}    200
    ${listed}=    Get Length    ${list_resp.json()}[items]

    IF    ${count} > 100
        Skip    More unread notifications than one page holds; the comparison cannot be made.
    END
    Should Be Equal As Integers    ${count}    ${listed}

An Unauthenticated Request Is Refused
    Create Anonymous API Session    alias=notif_anon_count
    ${resp}=    GET On Session    notif_anon_count    ${UNREAD_COUNT_API}    expected_status=any
    Should Be True    ${resp.status_code} in (401, 403)
