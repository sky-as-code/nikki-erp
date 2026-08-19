*** Settings ***
Documentation     Searching Vendor Profiles.
...
...               Two filters carry real weight. Filtering by party_id is how Purchase answers
...               "is this party a vendor?" — the profile's existence IS the answer, which is
...               the whole reason this is a table rather than a flag on the party. Filtering
...               by status is how a user picks from the suppliers actually orderable, since
...               only "active" may be named on a new order.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Vendor Profile Under Test
Test Tags         contacts    vendor_profile    search


*** Variables ***
${VENDOR_PROFILE_SCHEMA}    ${CONTACTS_SCHEMA_DIR}/vendor_profile.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${VENDOR_PROFILE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 0, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${VENDOR_PROFILE_SCHEMA}    size=7    page=0

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${VENDOR_PROFILE_SCHEMA}    size=50    page=99    item_count=0

Search By Party Succeeds
    [Documentation]    "Is this party a vendor, and on what terms?" — the query Purchase makes
    ...    through its port before accepting a vendor_id on an order.
    ${graph}=    Set Variable    {"if":["party_id","=","${PARTY_ID}"]}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': $graph, 'size': 100} }}
    Response Should Be Search Success    ${resp}    ${VENDOR_PROFILE_SCHEMA}    size=100    page=0
    Search Results Should Contain Id    ${resp}    ${VENDOR_PROFILE_ID}
    ${items}=    Set Variable    ${resp.json()}[items]
    FOR    ${item}    IN    @{items}
        Should Be Equal    ${item}[party_id]    ${PARTY_ID}
        ...    msg=A party-filtered search must not return another party's vendor profile
    END

Search By Status Succeeds
    [Documentation]    The listing a user picking a supplier sees. Every row must really carry
    ...    the requested status — a filter that silently widened would offer suspended vendors
    ...    for selection.
    ${graph}=    Set Variable    {"if":["status","=","active"]}
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'graph': $graph, 'size': 100} }}
    Response Should Be Search Success    ${resp}    ${VENDOR_PROFILE_SCHEMA}    size=100    page=0
    ${items}=    Set Variable    ${resp.json()}[items]
    FOR    ${item}    IN    @{items}
        Should Be Equal    ${item}[status]    active
        ...    msg=A status-filtered search must not return vendors of another status
    END

Search With Columns Succeeds
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'fields': ['party_id', 'status']} }}
    Response Status Should Be    ${resp}    200

Search With Nonexist Column Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}
    ...    params=${{ {'org_id': $CONTACTS_ORG_ID, 'fields': ['status', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
