*** Settings ***
Documentation     Notification Inbox - the NDJSON stream, at its contract edges (BR-FS 3,
...               BR-FS 4, BR-FS 14).
...
...               These tests deliberately do NOT read the body, and they cannot go through
...               RequestsLibrary to avoid it: every one of its keywords logs the response,
...               and logging reads the body, which on an endpoint designed never to finish
...               blocks until the test times out. StreamProbe makes the request directly
...               and closes it after the headers.
...
...               What one request can prove about an endless response is its status line
...               and its headers: that it opens, what it declares itself to be, who it
...               refuses, and what it must not change. The body's own contract — line
...               framing across chunk boundaries, heartbeats, dedupe by sequence, after_seq
...               resume — is covered by the micro-app's stream tests, which exercise
...               exactly that against a stubbed response.
Library           ${CURDIR}/../../nikki_api/StreamProbe.py
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    inbox    stream


*** Test Cases ***
The Stream Opens And Declares Itself As NDJSON
    [Documentation]    BR-FS 3. The content type is what tells a client to parse line by
    ...    line rather than waiting for a document that never arrives.
    ${stream}=    Probe Stream
    Should Be Equal As Integers    ${stream}[status_code]    200
    Should Contain    ${stream}[headers][Content-Type]    application/x-ndjson

The Stream Declares Itself Uncacheable And Untransformable
    [Documentation]    BR-FS 3. no-transform is the load-bearing half: a proxy that
    ...    re-encoded or buffered the body to "help" would turn a live stream into a
    ...    response that arrives all at once, or not until it ends.
    ${stream}=    Probe Stream
    Should Be Equal As Integers    ${stream}[status_code]    200
    Should Contain    ${stream}[headers][Cache-Control]    no-cache
    Should Contain    ${stream}[headers][Cache-Control]    no-transform

The Stream Declares No Content Length
    [Documentation]    BR-FS 3. A fixed length on a response that never ends would make
    ...    every client wait for bytes that are not coming.
    ${stream}=    Probe Stream
    Should Be Equal As Integers    ${stream}[status_code]    200
    Dictionary Should Not Contain Key    ${stream}[headers]    Content-Length

The Stream Accepts An After Seq Cursor
    [Documentation]    BR-FS 10. A reconnecting client passes the last sequence it
    ...    processed; the server replays what came after it before going live.
    ${stream}=    Probe Stream    after_seq=1
    Should Be Equal As Integers    ${stream}[status_code]    200

An Unauthenticated Stream Request Is Refused
    [Documentation]    AC-FS-02. The stream carries somebody's private correspondence, so
    ...    an anonymous caller must never get one open.
    Ensure Notification Org
    ${stream}=    Open Stream    ${API_HOST}${STREAM_API}    token=${EMPTY}
    ...    client_cert=${CLIENT_CERT}    client_key=${CLIENT_KEY}    verify=${SSL_VERIFY}
    ...    org_id=${NOTIF_ORG_ID}
    Should Be True    ${stream}[status_code] in (401, 403)

Opening The Stream Does Not Mark Anything Read
    [Documentation]    AC-FS-18, BR-FS 14. Receiving a notification is not reading it.
    ...    Only an explicit mark-read may change read_at — otherwise a background tab would
    ...    silently clear somebody's unread list.
    ${before}=    Get Unread Count
    Response Status Should Be    ${before}    200
    ${was}=    Set Variable    ${before.json()}[count]

    ${stream}=    Probe Stream
    Should Be Equal As Integers    ${stream}[status_code]    200

    ${after}=    Get Unread Count
    Response Status Should Be    ${after}    200
    Should Be Equal As Integers    ${after.json()}[count]    ${was}
    ...    msg=Opening a stream must not change read state


*** Keywords ***
Probe Stream
    [Documentation]    Opens the stream as the signed-in user and returns its status and
    ...    headers, having closed the connection without reading the body.
    [Arguments]    &{query}
    Ensure Notification Org
    Ensure Logged In
    ${stream}=    Open Stream    ${API_HOST}${STREAM_API}    token=${API_ACCESS_TOKEN}
    ...    client_cert=${CLIENT_CERT}    client_key=${CLIENT_KEY}    verify=${SSL_VERIFY}
    ...    org_id=${NOTIF_ORG_ID}    &{query}
    RETURN    ${stream}
