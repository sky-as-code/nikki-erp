*** Settings ***
Documentation     Creating a Vendor Profile. Saves the profile the later suites operate on.
...
...               The one-per-party-per-org rule is the whole reason this is a table rather
...               than columns on the party, so it is tested here rather than assumed.
Library           Collections
Library           RequestsLibrary
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Party Under Test
Test Tags         contacts    vendor_profile    create


*** Test Cases ***
Create With Required Fields Succeeds
    [Documentation]    party_id, org_id and status are the whole of what a vendor needs to
    ...    exist. Everything a purchase order defaults from is optional, because a supplier is
    ...    routinely approved before its payment terms are agreed.
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'status': 'active', 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${VENDOR_PROFILE_ID}    ${id}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}

Create With Full Payload Succeeds
    [Documentation]    Every optional field at once, on a party of its own so it does not
    ...    collide with the profile under test.
    ${party_id}    ${party_etag}=    Create Party    Robot Vendor Full    company
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $party_id, 'status': 'proposed', 'org_id': $CONTACTS_ORG_ID, 'status_reason': 'Awaiting compliance review', 'payment_terms': 'Net 30', 'lead_time_days': 14, 'note': 'Robot full vendor'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${id}
    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/vendor_profile.json    200
    ${item}=    Set Variable    ${resp.json()}[item]
    Should Be Equal    ${item}[payment_terms]    Net 30
    Should Be Equal As Integers    ${item}[lead_time_days]    14
    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${VENDOR_PROFILE_API}/${id}    expected_status=any
    ...    AND    DELETE On Session    api    ${PARTY_API}/${party_id}    expected_status=any

Status Defaults To Proposed
    [Documentation]    A profile created without a status is NOT orderable. Defaulting to
    ...    active would let an unqualified supplier be ordered from by omission — the one
    ...    default where the safe value and the convenient value differ.
    ${party_id}    ${party_etag}=    Create Party    Robot Vendor Default    company
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $party_id, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${id}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[item][status]    proposed
    [Teardown]    Run Keywords
    ...    DELETE On Session    api    ${VENDOR_PROFILE_API}/${id}    expected_status=any
    ...    AND    DELETE On Session    api    ${PARTY_API}/${party_id}    expected_status=any

Create Second Profile For Same Party And Org Fails
    [Documentation]    The 1-1 rule, and the reason this resource is a sidecar table. A second
    ...    row would give one supplier two sets of payment terms with no way to say which
    ...    applies to an order naming that vendor.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'status': 'active', 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    party_id    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    party_id    org_id

Create With Invalid Status Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'status': 'preferred', 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A status outside the four qualification states must not be accepted

Create With Negative Lead Time Fails
    [Documentation]    lead_time_days is a duration, so it has no meaningful negative value —
    ...    and a negative one would default a purchase order's expected arrival into the past.
    [Tags]    negative
    ${party_id}    ${party_etag}=    Create Party    Robot Vendor Negative    company
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $party_id, 'status': 'active', 'org_id': $CONTACTS_ORG_ID, 'lead_time_days': -1} }}
    ...    expected_status=any
    Response Should Be Invalid Number Range Error    ${resp}    lead_time_days
    [Teardown]    DELETE On Session    api    ${PARTY_API}/${party_id}    expected_status=any

Create With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'status': 'active', 'org_id': $CONTACTS_ORG_ID, 'discount_rate': 5} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    discount_rate

Create With Malformed Payload Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    data=not-a-json-object    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}
