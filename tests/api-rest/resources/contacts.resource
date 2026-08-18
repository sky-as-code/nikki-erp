*** Settings ***
Documentation     Contacts-module helpers shared by the party, comm channel and relationship
...               suites. The three resources form a dependency chain — a comm channel and a
...               relationship both hang off a party, and a relationship needs two — so each
...               fixture keyword is lazy and pulls in the ones it depends on. That keeps
...               every suite file runnable standalone.
...
...               Paths are the dynamic resource engine's: /v1/contacts/{schema_name}, with
...               the schema name spelled exactly as the Go constant. The engine asserts
...               permissions using the schema name as the resource code, so a path that
...               drifts from the schema is refused rather than 404'd, which reads as a
...               permission problem and not a wrong URL.
...
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${PARTY_API}                /v1/contacts/contacts_party
${COMM_CHANNEL_API}         /v1/contacts/contacts_comm_channel
${RELATIONSHIP_API}         /v1/contacts/contacts_relationship
${VENDOR_PROFILE_API}       /v1/contacts/contacts_vendor_profile
${CONTACTS_SCHEMA_DIR}      ${CURDIR}/schemas/contacts


*** Keywords ***
Ensure Contacts Org
    [Documentation]    Party and comm channel both declare org_id required-for-create.
    ...    ORG_ID from the environment wins when set; otherwise the first organization the
    ...    API returns is used, which keeps the suite app-agnostic instead of hardcoding a
    ...    seeded ULID.
    ...
    ...    Note contacts_relationship carries no org_id — it is scoped through the parties it
    ...    joins. That asymmetry is deliberate in the schema, so nothing here supplies one.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${CONTACTS_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${CONTACTS_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to own the Contacts fixtures; set ORG_ID in the environment
    Set Global Variable    ${CONTACTS_ORG_ID}    ${items}[0][id]

Unique Tax Id
    [Documentation]    tax_id carries no unique constraint — see the create suite for why —
    ...    but the value is still made unique per call so that a search asserting one hit
    ...    cannot be crossed by another test's leftover row. Kept short: the column is a
    ...    SHORT_NAME.
    [Arguments]    ${words}=tax
    ${suffix}=    Unique Suffix
    RETURN    ${words}${suffix}

Unique Website
    [Documentation]    website is data_type `url`, so a bare unique token would be rejected as
    ...    malformed. This wraps the token in a well-formed URL.
    ...
    ...    The value is still made unique per call even though the column no longer carries a
    ...    unique constraint, so that a search asserting one hit cannot be crossed by another
    ...    test's leftover row.
    [Arguments]    ${words}=party
    ${suffix}=    Unique Suffix
    RETURN    https://${words}${suffix}.example.com

# --- Party ------------------------------------------------------------------------------

Create Party
    [Documentation]    Creates a party of the given type and returns (id, etag). Callers that
    ...    need their own party — the relationship suite consumes two — use this instead of
    ...    the shared "under test" one.
    [Arguments]    ${words}=Robot Party    ${type}=individual
    Ensure Contacts Org
    ${name}=    Unique Display Name    ${words}
    ${resp}=    POST On Session    api    ${PARTY_API}
    ...    json=${{ {'display_name': $name, 'type': $type, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Party Under Test
    [Documentation]    The party the party suite exercises and that the comm channel suite
    ...    hangs its channels off. Created as an individual, the schema default.
    ${id}=    Get Variable Value    ${PARTY_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Party    Robot Party
    Set Global Variable    ${PARTY_ID}    ${id}
    Set Global Variable    ${PARTY_ETAG}    ${etag}

Ensure Target Party Under Test
    [Documentation]    A second party, for the rules that need two: a relationship runs
    ...    between a source and a target, and pointing one at itself is a case the schema
    ...    permits but the business does not.
    ${id}=    Get Variable Value    ${TARGET_PARTY_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Party    Robot Target Party    company
    Set Global Variable    ${TARGET_PARTY_ID}    ${id}
    Set Global Variable    ${TARGET_PARTY_ETAG}    ${etag}

# --- Comm channel -----------------------------------------------------------------------

Create Comm Channel
    [Documentation]    Adds one channel to a party and returns (id, etag). `value` is the
    ...    plain-text form; `value_json` carries the structured form a postal address needs
    ...    and is left unset here.
    [Arguments]    ${party_id}    ${type}=email    ${value}=${EMPTY}
    Ensure Contacts Org
    IF    not $value
        ${value}=    Unique Email
    END
    ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
    ...    json=${{ {'party_id': $party_id, 'type': $type, 'value': $value, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Comm Channel Under Test
    [Documentation]    The channel the comm_channel suite exercises, on the party under test.
    ${id}=    Get Variable Value    ${COMM_CHANNEL_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Party Under Test
    ${id}    ${etag}=    Create Comm Channel    ${PARTY_ID}    email
    Set Global Variable    ${COMM_CHANNEL_ID}    ${id}
    Set Global Variable    ${COMM_CHANNEL_ETAG}    ${etag}

# --- Relationship -----------------------------------------------------------------------

Create Relationship
    [Documentation]    Joins two parties and returns (id, etag). No org_id: the relationship
    ...    schema carries none, and sending one would be a nonexistent field.
    [Arguments]    ${party_id}    ${target_party_id}    ${type}=employee
    ${resp}=    POST On Session    api    ${RELATIONSHIP_API}
    ...    json=${{ {'party_id': $party_id, 'target_party_id': $target_party_id, 'type': $type} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Relationship Under Test
    [Documentation]    The relationship the relationship suite exercises, joining the two
    ...    parties under test.
    ${id}=    Get Variable Value    ${RELATIONSHIP_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Party Under Test
    Ensure Target Party Under Test
    ${id}    ${etag}=    Create Relationship    ${PARTY_ID}    ${TARGET_PARTY_ID}    employee
    Set Global Variable    ${RELATIONSHIP_ID}    ${id}
    Set Global Variable    ${RELATIONSHIP_ETAG}    ${etag}

# --- Bulk seeding -----------------------------------------------------------------------

Ensure Seeded Parties
    [Documentation]    Bulk-creates sample parties for Search/Exists once per execution.
    ...    Every 10th carries "Lead" in the display name so content-dependent search filters
    ...    match even on an otherwise empty database. Left in the DB by design.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_PARTY_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Contacts Org
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Party ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Party ${index}
        END
        ${resp}=    POST On Session    api    ${PARTY_API}
        ...    json=${{ {'display_name': $name, 'type': 'individual', 'org_id': $CONTACTS_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_PARTY_IDS}    ${ids}

Ensure Seeded Comm Channels
    [Documentation]    Bulk-creates sample channels on the party under test for
    ...    Search/Exists. Types are cycled so a filter on `type` has more than one bucket.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_COMM_CHANNEL_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Party Under Test
    ${types}=    Create List    email    phone    zalo    facebook    post
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${slot}=    Evaluate    $index % 5
        ${type}=    Set Variable    ${types}[${slot}]
        IF    '${type}' == 'email'
            ${value}=    Unique Email
        ELSE
            ${suffix}=    Unique Suffix
            ${value}=    Set Variable    seed${suffix}
        END
        ${resp}=    POST On Session    api    ${COMM_CHANNEL_API}
        ...    json=${{ {'party_id': $PARTY_ID, 'type': $type, 'value': $value, 'org_id': $CONTACTS_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_COMM_CHANNEL_IDS}    ${ids}

# --- Vendor profile ---------------------------------------------------------------------

Create Vendor Profile
    [Documentation]    Creates a vendor profile for the given party and returns (id, etag).
    ...
    ...    Only party_id, org_id and status are sent. The three defaults a purchase order reads
    ...    — currency, payment terms and lead time — are all optional, because a vendor is
    ...    frequently recorded before its terms are agreed, and a fixture that always supplied
    ...    them would never exercise that.
    [Arguments]    ${party_id}    ${status}=active
    Ensure Contacts Org
    ${resp}=    POST On Session    api    ${VENDOR_PROFILE_API}
    ...    json=${{ {'party_id': $party_id, 'status': $status, 'org_id': $CONTACTS_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Ensure Vendor Profile Under Test
    [Documentation]    The profile the vendor suite exercises, hung off the shared party.
    ...    Created active, since that is the only status a purchase order may name and so the
    ...    one most tests want.
    ${id}=    Get Variable Value    ${VENDOR_PROFILE_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Party Under Test
    ${id}    ${etag}=    Create Vendor Profile    ${PARTY_ID}
    Set Global Variable    ${VENDOR_PROFILE_ID}    ${id}
    Set Global Variable    ${VENDOR_PROFILE_ETAG}    ${etag}

# --- Teardown ---------------------------------------------------------------------------

Delete Contacts Fixture
    [Documentation]    Best-effort delete of one record, clearing the global that held it so
    ...    a later suite cannot reuse a dead id.
    [Arguments]    ${api}    ${variable_name}
    ${id}=    Get Variable Value    ${${variable_name}}    ${EMPTY}
    IF    $id
        DELETE On Session    api    ${api}/${id}    expected_status=any
        Set Global Variable    ${${variable_name}}    ${EMPTY}
    END

Delete Contacts Seed Data
    [Documentation]    Best-effort removal of the fixtures the Contacts suites created on the
    ...    fly, referencing-before-referenced. A comm channel, a relationship and a vendor
    ...    profile all point at a party, so all three go before either party does.
    ...
    ...    Every one of those FKs is ON DELETE CASCADE, so the parties would take their children
    ...    with them — but the order is written out anyway, because relying on the cascade would
    ...    hide a schema change that removed it.
    Delete Contacts Fixture    ${VENDOR_PROFILE_API}    VENDOR_PROFILE_ID
    Delete Contacts Fixture    ${RELATIONSHIP_API}    RELATIONSHIP_ID
    Delete Contacts Fixture    ${COMM_CHANNEL_API}    COMM_CHANNEL_ID
    Delete Contacts Fixture    ${PARTY_API}    TARGET_PARTY_ID
    Delete Contacts Fixture    ${PARTY_API}    PARTY_ID
