*** Settings ***
Documentation     Creating Parties. The first test saves the party under test
...               (${PARTY_ID}/${PARTY_ETAG}) consumed by the later suites and deleted last
...               by 08_delete.robot.
Resource          resources/contacts.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Contacts Org
Test Tags         contacts    party    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Party
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PARTY_ID}    ${id}
    Set Global Variable    ${PARTY_ETAG}    ${etag}
    Set Global Variable    ${PARTY_NAME}    ${name}

Create Company Party Succeeds
    [Documentation]    A company is the party type a vendor is: contacts_vendor_profile hangs
    ...    off one of these, so the type must be creatable in its own right.
    ${name}=    Unique Display Name    Robot Company Party
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'company', 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PARTY_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[type]    company
    DELETE On Session    api    ${PARTY_API}/${id}    expected_status=any

Create With All Optional Fields Succeeds
    [Documentation]    website and avatar_url are data_type `url`, not plain strings; this
    ...    pins that well-formed URLs are accepted and stored on the record.
    ${name}=    Unique Display Name    Robot Full Party
    ${tax_id}=    Unique Tax Id    full
    ${website}=    Unique Website    fullparty
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'company', 'org_id': $CONTACTS_ORG_ID, 'legal_name': 'Robot Legal Name', 'legal_address': '1 Robot Street', 'tax_id': $tax_id, 'job_position': 'Buyer', 'title': 'Mr', 'note': 'A robot party', 'website': $website, 'avatar_url': 'https://example.com/avatar.png'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PARTY_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[tax_id]    ${tax_id}
    Should Be Equal    ${item}[website]    ${website}
    DELETE On Session    api    ${PARTY_API}/${id}    expected_status=any

Create Defaults Type To Individual
    [Documentation]    `type` carries default_value "individual", so omitting it must not be
    ...    a missing-field error.
    ${name}=    Unique Display Name    Robot Default Type
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PARTY_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${CONTACTS_SCHEMA_DIR}/party.json    200
    Should Be Equal    ${item}[type]    individual
    DELETE On Session    api    ${PARTY_API}/${id}    expected_status=any

Create With Invalid Type Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Type
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'supplier', 'org_id': $CONTACTS_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A party type outside the enum must not be accepted on create

Create With Malformed Website Fails
    [Documentation]    website is data_type `url`; no dedicated assertion keyword pins the
    ...    exact error key for this contract, so this only asserts the create is rejected.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Website Party
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID, 'website': 'not a url'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A malformed website URL must not be accepted on create

Create With Duplicate Tax Id Succeeds
    [Documentation]    tax_id carries NO unique constraint, and that is deliberate rather than an
    ...    oversight — this test exists so the decision is visible instead of being rediscovered.
    ...
    ...    It used to be globally unique, which was a multi-tenancy bug: the first org to record a
    ...    supplier locked every other org out of recording its own. The obvious fix, unique per
    ...    org, is not expressible here. The framework's partial unique emits a companion
    ...    "UNIQUE (org_id) WHERE tax_id IS NULL" index, which would allow only ONE party per org
    ...    with no tax id — and every individual contact has none.
    ...
    ...    So duplicates are accepted by the database and detecting them belongs to the
    ...    application layer, where it can warn instead of refusing a write. That is better
    ...    behaviour regardless: branches of one group legitimately share a registration number.
    ...
    ...    If a duplicate check is added later it should make this test fail, not be bolted on
    ...    beside it.
    ${tax_id}=    Unique Tax Id    dup
    ${first}=    Unique Display Name    Robot Tax Holder
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $first, 'type': 'company', 'org_id': $CONTACTS_ORG_ID, 'tax_id': $tax_id} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${second}=    Unique Display Name    Robot Tax Duplicate
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $second, 'type': 'company', 'org_id': $CONTACTS_ORG_ID, 'tax_id': $tax_id} }}
    ${second_id}    ${second_etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${PARTY_API}/${second_id}    expected_status=any
    DELETE On Session    api    ${PARTY_API}/${id}    expected_status=any

Create Two Parties Without Tax Id Succeeds
    [Documentation]    The case that ruled out the partial unique. Most contacts have no tax id at
    ...    all, so an org must be able to hold any number of them. Under a partial unique the
    ...    second create here fails with a duplicate-key error naming org_id — a message that
    ...    would send someone looking in entirely the wrong place.
    ${first}=    Unique Display Name    Robot Untaxed One
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $first, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID} }}
    ${first_id}    ${first_etag}=    Response Should Be Create Success    ${resp}
    ${second}=    Unique Display Name    Robot Untaxed Two
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $second, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID} }}
    ${second_id}    ${second_etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${PARTY_API}/${second_id}    expected_status=any
    DELETE On Session    api    ${PARTY_API}/${first_id}    expected_status=any    expected_status=any

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PARTY_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    display_name    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    data={ "display_name": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
