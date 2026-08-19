*** Settings ***
Documentation     Creating Relationships. The first test saves the relationship under test
...               (${RELATIONSHIP_ID}/${RELATIONSHIP_ETAG}) consumed by the later suites and
...               deleted last by 08_delete.robot.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Party Under Test
...               AND    Ensure Target Party Under Test
Test Tags         contacts    relationship    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'target_party_id': $TARGET_PARTY_ID, 'type': 'employee'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${RELATIONSHIP_ID}    ${id}
    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}

Create Subsidiary Relationship Succeeds
    [Documentation]    subsidiary is the company-to-company kind, the one that matters for a
    ...    vendor group: a purchase order names one legal entity, and its parent is a
    ...    different party.
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $TARGET_PARTY_ID, 'target_party_id': $PARTY_ID, 'type': 'subsidiary', 'note': 'Robot subsidiary'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${RELATIONSHIP_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/relationship.json    200
    Should Be Equal    ${item}[type]    subsidiary
    DELETE On Session    api    ${RELATIONSHIP_API}/${id}    expected_status=any

Create With Invalid Type Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'target_party_id': $TARGET_PARTY_ID, 'type': 'supplier'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A relationship type outside the enum must not be accepted on create

Create With Not Found Source Party Fails
    [Documentation]    Both ends carry a real FK to contacts_parties.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $NOT_FOUND_ID, 'target_party_id': $TARGET_PARTY_ID, 'type': 'employee'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A relationship must not be created from a nonexistent party

Create With Not Found Target Party Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'target_party_id': $NOT_FOUND_ID, 'type': 'employee'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A relationship must not be created to a nonexistent party

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    party_id    target_party_id    type

Create With Org Id Fails
    [Documentation]    contacts_relationship declares no org_id, so sending one is a
    ...    nonexistent field rather than a silently ignored extra. This is the negative half
    ...    of "Schema Declares No Org Id" in 01: the two together stop a caller from assuming
    ...    the column exists because its siblings have it.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'target_party_id': $TARGET_PARTY_ID, 'type': 'employee', 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    data={ "type": "employee",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'target_party_id': $TARGET_PARTY_ID, 'type': 'employee', 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
