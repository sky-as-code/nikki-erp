*** Settings ***
Documentation     Updating Parties. The success cases run first (they consume and rotate the
...               saved etag); negatives follow.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Party Under Test
Test Tags         contacts    party    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Party
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    json=${{ {'display_name': $name, 'etag': $PARTY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PARTY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PARTY_ETAG}    ${etag}

Update Website Succeeds
    [Documentation]    website is an ordinary mutable optional field, same as any other.
    ${website}=    Unique Website    updated
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    json=${{ {'website': $website, 'etag': $PARTY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PARTY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PARTY_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[website]    ${website}
    Set Global Variable    ${PARTY_ETAG}    ${item}[etag]

Update Type Succeeds
    [Documentation]    A contact first recorded as an individual can turn out to be a
    ...    business. `type` is a plain mutable field, not a lifecycle state, so this is an
    ...    ordinary edit.
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    json=${{ {'type': 'company', 'etag': $PARTY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PARTY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PARTY_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PARTY_API}/${PARTY_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[type]    company
    Set Global Variable    ${PARTY_ETAG}    ${item}[etag]

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Party
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${PARTY_ID}
    ...    json=${{ {'display_name': $name, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PARTY_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PARTY_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PARTY_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PARTY_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
