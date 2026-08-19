*** Settings ***
Documentation     The Vendor Profile schema is served by the dynamic resource engine at
...               /v1/contacts/contacts_vendor_profile.
...
...               This resource replaced purchase_vendor, which lived in the Purchase module
...               and held nothing but a status — no name, no tax id, no address, and no link
...               to the contact it described. The requirement says outright that Purchase must
...               not own the vendor, so the concept moved here and the fields it was missing
...               come from the party it hangs off.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Test Tags         contacts    vendor_profile    schema


*** Test Cases ***
Get Vendor Profile Model Schema
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    party_id
    Dictionary Should Contain Key    ${fields}    org_id
    Dictionary Should Contain Key    ${fields}    status
    Dictionary Should Contain Key    ${fields}    status_reason
    Dictionary Should Contain Key    ${fields}    default_currency_id
    Dictionary Should Contain Key    ${fields}    payment_terms
    Dictionary Should Contain Key    ${fields}    lead_time_days

Status Enum Is Exactly The Four Qualification States
    [Documentation]    Pinned as an exact set rather than a subset. A fifth value added without
    ...    a matching decision in Purchase — which treats only "active" as orderable — would
    ...    silently be unorderable with nothing saying so.
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${status}=    Set Variable    ${resp.json()}[fields][status]
    ${values}=    Set Variable    ${status}[data_type][values]
    Lists Should Be Equal    ${values}    ${{ ['proposed', 'active', 'suspended', 'blacklisted'] }}
    ...    ignore_order=${True}

Schema Declares No Vendor Name Or Tax Id
    [Documentation]    The profile deliberately carries none of the contact's own facts. They
    ...    live on the party, and duplicating them here would let the two disagree about who
    ...    the supplier is.
    ${resp}=    GET On Session    api    ${VENDOR_PROFILE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Not Contain Key    ${fields}    display_name
    Dictionary Should Not Contain Key    ${fields}    legal_name
    Dictionary Should Not Contain Key    ${fields}    tax_id
    Dictionary Should Not Contain Key    ${fields}    legal_address
