*** Settings ***
Documentation     Archiving the Comm Channel under test, rotating the saved etag. The channel
...               is unarchived again so the later suites see it live.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
Test Tags         contacts    comm_channel    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}/archived
    ...    json=${{ {'etag': $COMM_CHANNEL_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${COMM_CHANNEL_ETAG}
    IF    $etag is not None    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}

Archived Channel Is Still Readable
    [Documentation]    A retired phone number stays on the record: correspondence sent to it
    ...    still happened, so archiving hides it from pickers without erasing the history.
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/comm_channel.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Archiving A Channel Leaves Its Party Live
    [Documentation]    The channel is the archivable thing, not the contact. Withdrawing one
    ...    way of reaching someone must not withdraw the someone.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[is_archived]    ${False}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}/archived
    ...    json=${{ {'etag': $COMM_CHANNEL_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${COMM_CHANNEL_ETAG}
    IF    $etag is not None    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $COMM_CHANNEL_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
