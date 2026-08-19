*** Settings ***
Documentation     Updating Comm Channels. The success cases run first (they consume and
...               rotate the saved etag); negatives follow.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
Test Tags         contacts    comm_channel    update


*** Test Cases ***
Update Value Succeeds
    ${value}=    Unique Email
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    json=${{ {'value': $value, 'etag': $COMM_CHANNEL_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${COMM_CHANNEL_ETAG}
    IF    $etag is not None    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/comm_channel.json    200
    Should Be Equal    ${item}[value]    ${value}
    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${item}[etag]

Update Note Succeeds
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    json=${{ {'note': 'Preferred contact', 'etag': $COMM_CHANNEL_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${COMM_CHANNEL_ETAG}
    IF    $etag is not None    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    ...    json=${{ {'note': 'Stale', 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $COMM_CHANNEL_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${COMM_CHANNEL_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $COMM_CHANNEL_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
