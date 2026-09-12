*** Settings ***
Documentation     Notification - Get model schema, for all three resources.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
Test Tags         notification    schema


*** Test Cases ***
Get Notification Model Schema
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Get Recipient Model Schema
    ${resp}=    GET On Session    api    ${RECIPIENT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Get Delivery Model Schema
    ${resp}=    GET On Session    api    ${DELIVERY_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Notification Carries No Read State
    [Documentation]    Read state belongs to the recipient and nowhere else (BR 26). A
    ...    read column on the notification could disagree with it, and a notification sent
    ...    to three people has three answers rather than one.
    ${resp}=    GET On Session    api    ${NOTIFICATION_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json()['fields'].keys())
    Should Not Contain    ${names}    read_at
    Should Not Contain    ${names}    is_read

Recipient Carries The Stream Sequence And Read State
    [Documentation]    stream_seq is the durable order of one person's stream and the
    ...    cursor a disconnected client replays from; read_at is where "read" lives.
    ${resp}=    GET On Session    api    ${RECIPIENT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json()['fields'].keys())
    Should Contain    ${names}    stream_seq
    Should Contain    ${names}    read_at
    Should Contain    ${names}    recipient_user_id

Neither Resource Is Archivable
    [Documentation]    Neither has an archive action (BR 6, BR 8): a notification's only
    ...    lifecycle is expiring, and expiry neither deletes it nor hides it from history.
    FOR    ${api}    IN    ${NOTIFICATION_API}    ${RECIPIENT_API}    ${DELIVERY_API}
        ${resp}=    GET On Session    api    ${api}/meta/schema
        Response Status Should Be    ${resp}    200
        ${names}=    Evaluate    list($resp.json()['fields'].keys())
        Should Not Contain    ${names}    is_archived
    END
