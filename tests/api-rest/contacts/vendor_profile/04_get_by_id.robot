*** Settings ***
Documentation     Reading a single Vendor Profile.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    get


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/vendor_profile.json    200
    Should Be Equal    ${item}[party_id]    ${PARTY_ID}
    Should Be Equal    ${item}[org_id]    ${CONTACTS_ORG_ID}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${item}[etag]

Get Returns The Defaults A Purchase Order Reads
    [Documentation]    The three fields Purchase pulls off a vendor to default an order.
    ...    All are optional, so this asserts they are PRESENT in the response — a caller must
    ...    be able to tell "not stated" from a value, and a field omitted entirely from the
    ...    payload cannot express either.
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}[item]
    Dictionary Should Contain Key    ${item}    default_currency_id
    Dictionary Should Contain Key    ${item}    payment_terms
    Dictionary Should Contain Key    ${item}    lead_time_days

Get With Columns Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    params=${{ {'fields': ['party_id', 'status', 'payment_terms']} }}
    Response Status Should Be    ${resp}    200

Get With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${VENDOR_PROFILE_ID}
    ...    params=${{ {'fields': ['status', 'bla_bla_field']} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
