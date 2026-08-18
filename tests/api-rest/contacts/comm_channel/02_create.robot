*** Settings ***
Documentation     Creating Comm Channels. The first test saves the channel under test
...               (${COMM_CHANNEL_ID}/${COMM_CHANNEL_ETAG}) consumed by the later suites and
...               deleted last by 08_delete.robot.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Party Under Test
Test Tags         contacts    comm_channel    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${value}=    Unique Email
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'type': 'email', 'value': $value, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${COMM_CHANNEL_ID}    ${id}
    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}

Create With Structured Value Succeeds
    [Documentation]    value_json is data_type `jsonmap`, which is what a postal address needs
    ...    — a single `value` string cannot hold street, ward and city separately.
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'type': 'post', 'value_json': {'street': '1 Robot Street', 'city': 'Da Nang'}, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${COMM_CHANNEL_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/comm_channel.json    200
    Should Be Equal    ${item}[value_json][city]    Da Nang
    DELETE On Session    api    ${COMM_CHANNEL_API}/${id}    expected_status=any

Create Every Channel Type Succeeds
    [Documentation]    One party legitimately holds several channels of different kinds, and
    ...    nothing constrains the combination — so every enum value must be creatable against
    ...    the same party.
    FOR    ${type}    IN    phone    zalo    facebook    post
        ${suffix}=    Unique Suffix
        ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
        ...    json=${{ {'party_id': $PARTY_ID, 'type': $type, 'value': 'val' + $suffix, 'org_id': $CONTACTS_ORG_ID} }}
        ${id}    ${etag}=    Response Should Be Create Success    ${resp}
        DELETE On Session    api    ${COMM_CHANNEL_API}/${id}    expected_status=any
    END

Create With Invalid Type Fails
    [Tags]    negative
    ${suffix}=    Unique Suffix
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'type': 'telegram', 'value': 'val' + $suffix, 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A channel type outside the enum must not be accepted on create

Create With Not Found Party Fails
    [Documentation]    party_id carries a real FK to contacts_parties, so a channel cannot
    ...    hang off a party that does not exist.
    [Tags]    negative
    ${value}=    Unique Email
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $NOT_FOUND_ID, 'type': 'email', 'value': $value, 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A comm channel must not be created against a nonexistent party

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    party_id    type    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    data={ "type": "email",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${value}=    Unique Email
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $PARTY_ID, 'type': 'email', 'value': $value, 'org_id': $CONTACTS_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
