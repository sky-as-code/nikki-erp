*** Settings ***
Documentation     Deleting the Party under test — always the LAST suite, doubling as cleanup.
...               A party owns its comm channels and relationships through ON DELETE CASCADE,
...               so deleting one takes its children with it rather than being refused. That
...               is the opposite of the Inventory master data, where every FK is NO ACTION,
...               and it is why there is no "delete with referencing records fails" case here.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Party Under Test
Test Tags         contacts    party    delete


*** Test Cases ***
Delete Cascades To Comm Channels
    [Documentation]    A channel is meaningless without the party it belongs to, so the FK
    ...    cascades. Pinned on a throwaway party, because the party under test is deleted by
    ...    the next case and this must not depend on that ordering.
    ${party_id}    ${party_etag}=    Create Party    Robot Cascade Party
    ${channel_id}    ${channel_etag}=    Create Comm Channel    ${party_id}    email
    ${resp}=    DELETE On Session    api    ${PARTY_API}/${party_id}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${channel_id}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete Succeeds
    ${resp}=    DELETE On Session    api    ${PARTY_API}/${PARTY_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    party under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PARTY_API}/${PARTY_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PARTY_ID}    ${EMPTY}
    Set Global Variable    ${PARTY_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PARTY_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PARTY_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
