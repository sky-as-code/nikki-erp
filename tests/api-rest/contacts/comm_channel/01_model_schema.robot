*** Settings ***
Documentation     The Comm Channel schema is served by the dynamic resource engine. Its path
...               is a top-level resource, /v1/contacts/contacts_comm_channel, not nested
...               under a party as the hand-written API had it
...               (/v1/:org_id/contacts/parties/:party_id/channels). The engine addresses
...               every resource by its own schema name; the party is a field, not a path.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Test Tags         contacts    comm_channel    schema


*** Test Cases ***
Get Comm Channel Model Schema
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    party_id
    Dictionary Should Contain Key    ${fields}    type
    Dictionary Should Contain Key    ${fields}    value
    Dictionary Should Contain Key    ${fields}    value_json
    Dictionary Should Contain Key    ${fields}    org_id

Schema Declares The Channel Type Enum
    [Documentation]    The five channel kinds. `post` is the one that carries a structured
    ...    address in value_json rather than a single string in value.
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][type][data_type][values]
    Should Contain    ${values}    phone
    Should Contain    ${values}    zalo
    Should Contain    ${values}    facebook
    Should Contain    ${values}    email
    Should Contain    ${values}    post
    Length Should Be    ${values}    5
