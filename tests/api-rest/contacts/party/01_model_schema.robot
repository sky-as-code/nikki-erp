*** Settings ***
Documentation     The Party schema is served by the dynamic resource engine, same as the
...               other two Contacts resources. This suite is also the cutover check: the
...               engine path /v1/contacts/contacts_party replaced the hand-written
...               /v1/:org_id/contacts/parties, and the schema name in the URL must match
...               the Go constant byte for byte or the engine refuses the request.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Test Tags         contacts    party    schema


*** Test Cases ***
Get Party Model Schema
    ${resp}=    GET On Session    api    ${PARTY_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${PARTY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    display_name
    Dictionary Should Contain Key    ${fields}    type
    Dictionary Should Contain Key    ${fields}    tax_id
    Dictionary Should Contain Key    ${fields}    website
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares No Status Field
    [Documentation]    A party's lifecycle is archive alone. A status field would invite a
    ...    second, conflicting notion of "retired" — and vendor qualification status lives on
    ...    contacts_vendor_profile, not here, precisely so the two cannot be confused.
    ${resp}=    GET On Session    api    ${PARTY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a party shows as an opaque id when chosen as the vendor of a
    ...    purchase order.
    ${resp}=    GET On Session    api    ${PARTY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    display_name

Schema Declares The Party Type Enum
    [Documentation]    individual and company are the only two values, and the frontend
    ...    renders the field as a select off this list rather than a free-text box.
    ${resp}=    GET On Session    api    ${PARTY_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${type_field}=    Set Variable    ${resp.json()}[fields][type]
    ${values}=    Set Variable    ${type_field}[data_type][values]
    Should Contain    ${values}    individual
    Should Contain    ${values}    company
    Length Should Be    ${values}    2
