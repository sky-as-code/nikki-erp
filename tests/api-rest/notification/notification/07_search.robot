*** Settings ***
Documentation     Notification - Search, and the org scoping it must not be able to escape.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    search


*** Variables ***
${NOTIFICATION_SCHEMA}    ${NOTIFICATION_SCHEMA_DIR}/notification.json


*** Test Cases ***
Search Returns The Default Page
    Ensure Notification Org
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID} }}
    Response Status Should Be    ${resp}    200

Search With Paging Returns The Requested Page
    Ensure Notification Org
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID, 'page': 1, 'size': 5} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[size]    5
    Should Be Equal As Integers    ${resp.json()}[page]    1

Search By Severity Returns Only That Severity
    Ensure Notification Org
    ${graph}=    Set Variable    {"if": ["severity", "=", "info"]}
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID, 'graph': $graph, 'size': 20} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[severity]    info
    END

Every Result Belongs To The Requested Organization
    [Documentation]    AC14. The org predicate is applied by the server, and a listing that
    ...    returned another organization's notifications would leak business content — the
    ...    title and message are written for a specific audience.
    Ensure Notification Org
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID, 'fields': 'id,org_id', 'size': 50} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[org_id]    ${NOTIF_ORG_ID}
    END

A Client Graph Cannot Widen The Search Past Its Organization
    [Documentation]    The org predicate is ANDed on top of whatever graph the client sent,
    ...    never merged into it. Scoping that the request it scopes can override is not
    ...    scoping at all.
    Ensure Notification Org
    # A graph that on its own would match every notification in the table.
    ${wide}=    Set Variable    {"if": ["severity", "!=", "there_is_no_such_severity"]}
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}
    ...    params=${{ {'org_id': $NOTIF_ORG_ID, 'fields': 'id,org_id', 'graph': $wide, 'size': 50} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[org_id]    ${NOTIF_ORG_ID}
        ...    msg=A client graph must not widen the listing past the caller's organization
    END
