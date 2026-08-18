*** Settings ***
Documentation     Deleting the Comm Channel under test — always the LAST suite, doubling as
...               cleanup. A channel is a leaf: nothing references it, so it deletes cleanly
...               and there is no "delete with referencing records fails" case here.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Comm Channel Under Test
Test Tags         contacts    comm_channel    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}
    Response Should Be Delete Success    ${resp}    count=1

Deleting A Channel Leaves Its Party Live
    [Documentation]    The cascade runs one way only: a party takes its channels with it, but
    ...    removing a channel must not touch the contact it belonged to.
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    channel under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${COMM_CHANNEL_API}/${COMM_CHANNEL_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${COMM_CHANNEL_ID}    ${EMPTY}
    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${COMM_CHANNEL_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${COMM_CHANNEL_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
