*** Settings ***
Documentation     Updating the Vendor Profile under test. The engine's update is a PATCH, so a
...               payload names only the fields it changes.
Library           Collections
Library           RequestsLibrary
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    update


*** Test Cases ***
Update Terms Succeeds
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'payment_terms': 'Net 45', 'lead_time_days': 21} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ${item}=    Set Variable    ${resp.json()}[item]
    Should Be Equal    ${item}[payment_terms]    Net 45
    Should Be Equal As Integers    ${item}[lead_time_days]    21

Suspend With Reason Succeeds
    [Documentation]    Qualification status is an ordinary field in this phase, not a state
    ...    machine, so suspending is a plain update. If a transition ever grows a side effect
    ...    — notifying open orders, say — it becomes a domain action and this test should be
    ...    replaced rather than extended.
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'status': 'suspended', 'status_reason': 'Late deliveries'} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    Should Be Equal    ${resp.json()}[item][status]    suspended

Reactivate Succeeds
    [Documentation]    Restores the profile to active so the later suites see the status they
    ...    expect. Every transition here is reversible precisely because none has a side
    ...    effect.
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'status': 'active'} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}

Update Party Id Is Refused Or Ignored
    [Documentation]    party_id is no_update: moving a profile to another party would silently
    ...    reassign every purchase order that named the first one.
    ...
    ...    The engine's response to a no_update field is not uniform — some are refused with a
    ...    400 and some answered 200 with the field dropped — so this asserts the STORED VALUE
    ...    rather than the status code. A test pinned to the code would keep passing while the
    ...    behaviour changed underneath it, and it is the stored value that matters here.
    [Tags]    negative
    ${other_id}    ${other_etag}=    Create Party    Robot Vendor Reassign    company
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'party_id': $other_id} }}
    ...    expected_status=any
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[item][party_id]    ${PARTY_ID}
    ...    msg=party_id must not be reassignable through a plain update
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${resp.json()}[item][etag]
    [Teardown]    DELETE On Session    api    ${PARTY_API}/${other_id}    expected_status=any

Update With Stale Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': '1', 'payment_terms': 'Net 60'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'payment_terms': 'Net 60'} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Update With Invalid Status Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    json=${{ {'etag': $VENDOR_PROFILE_ETAG, 'status': 'preferred'} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A status outside the four qualification states must not be accepted
