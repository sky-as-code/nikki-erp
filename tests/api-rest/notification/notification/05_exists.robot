*** Settings ***
Documentation     Notification - Exists.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${id}=    Ensure Notification In Inbox
    ${resp}=    POST On Session    api    ${NOTIFICATION_API}/exists
    ...    json=${{ {'org_id': $NOTIF_ORG_ID, 'ids': [$id]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists Reports An Unknown Id As Not Existing
    Ensure Notification Org
    ${resp}=    POST On Session    api    ${NOTIFICATION_API}/exists
    ...    json=${{ {'org_id': $NOTIF_ORG_ID, 'ids': [$NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=0    not_existing=1

Exists Separates The Known From The Unknown
    [Documentation]    The point of the endpoint is answering for a batch: a client holding
    ...    a list of ids learns which of them still resolve, in one call rather than one per
    ...    id.
    ${id}=    Ensure Notification In Inbox
    ${fakes}=    Not Found Id List    3
    ${ids}=    Combine Lists    ${{ [$id] }}    ${fakes}
    ${resp}=    POST On Session    api    ${NOTIFICATION_API}/exists
    ...    json=${{ {'org_id': $NOTIF_ORG_ID, 'ids': $ids} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=3
