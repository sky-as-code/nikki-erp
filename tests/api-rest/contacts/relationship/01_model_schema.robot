*** Settings ***
Documentation     The Relationship schema is served by the dynamic resource engine at
...               /v1/contacts/contacts_relationship, replacing the nested
...               /v1/:org_id/contacts/parties/:party_id/relationships.
...
...               Relationship is the one Contacts resource carrying no org_id: it is scoped
...               through the two parties it joins, both of which are org-scoped themselves.
...               That asymmetry is pinned here rather than left implicit, because a later
...               "consistency" fix that adds org_id would change how every query filters.
Resource          resources/contacts.resource
Suite Setup       Create Authorized API Session
Test Tags         contacts    relationship    schema


*** Test Cases ***
Get Relationship Model Schema
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Dictionary Should Contain Key    ${fields}    party_id
    Dictionary Should Contain Key    ${fields}    target_party_id
    Dictionary Should Contain Key    ${fields}    type
    Dictionary Should Contain Key    ${fields}    note

Schema Declares No Org Id
    [Documentation]    Unlike party and comm channel, a relationship carries no org_id — see
    ...    the suite documentation. If this starts failing, the schema gained one and every
    ...    search in this suite needs the parameter.
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    org_id

Schema Declares The Relationship Type Enum
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${values}=    Set Variable    ${resp.json()}[fields][type][data_type][values]
    Should Contain    ${values}    employee
    Should Contain    ${values}    spouse
    Should Contain    ${values}    parent
    Should Contain    ${values}    sibling
    Should Contain    ${values}    emergency
    Should Contain    ${values}    subsidiary
    Length Should Be    ${values}    6
