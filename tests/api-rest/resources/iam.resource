*** Settings ***
Documentation     IAM-specific helpers shared by the IAM suites: the lazily-created
...               user-under-test, bulk sample data for Search/Exists, and the seed
...               user/group pair (created on the fly when the environment does not
...               provide SEED_USER_ID / SEED_GROUP_ID, deleted by Delete Seed Data).
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${USER_API}             /v1/iam/iam_user
${GROUP_API}            /v1/iam/iam_group
${ROLE_API}             /v1/iam/roles
${ENTITLEMENT_API}      /v1/iam/entitlements
${TEST_PERM_API}        /v1/iam/me/test-permissions
${TEMP_PASSWORD_API}    /v1/iam/passwords/passwordtmp
${IAM_SCHEMA_DIR}       ${CURDIR}/schemas/iam
# Records created by the permission suites, keyed by kind, removed by their teardown.
${PERM_FIXTURES}        ${None}
# Ids of the locale-ordering groups, removed by Delete Locale Group Fixtures.
${LOCALE_GROUP_IDS}     ${None}


*** Keywords ***
Ensure User Under Test
    [Documentation]    Creates the user that the whole iam/user suite exercises
    ...    (CREATE saves it, UPDATE/GET/ARCHIVE reuse it, DELETE removes it).
    ...    Lazy so any suite file can also run standalone.
    ${id}=    Get Variable Value    ${USER_ID}    ${EMPTY}
    IF    $id    RETURN
    ${email}=    Unique Email    robot.user
    ${name}=    Unique Display Name    Robot Test User
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${USER_ID}    ${id}
    Set Global Variable    ${USER_ETAG}    ${etag}
    Set Global Variable    ${USER_EMAIL}    ${email}

Ensure Seeded Users
    [Documentation]    Bulk-creates sample users for Search/Exists once per execution
    ...    (replaces Bruno's "Create test data - Loop 50"). Left in the DB by design.
    ...    Every 10th user carries "Lead" in the display name and every 5th-of-10 an
    ...    "Admin" name with a .com email, so the content-dependent search filters
    ...    match even on an otherwise empty database.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_USER_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead User ${index}
            ${email}=    Unique Email    seed.lead
        ELSE IF    $remainder == 5
            ${name}=    Unique Display Name    Seed Admin User ${index}
            ${email}=    Unique Email    seed.admin    domain=e2e.nikki.com
        ELSE
            ${name}=    Unique Display Name    Seed User ${index}
            ${email}=    Unique Email    seed.user
        END
        ${resp}=    POST On Session    api    ${USER_API}
        ...    json=${{ {'display_name': $name, 'email': $email} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_USER_IDS}    ${ids}

Ensure Seed User
    [Documentation]    A pre-existing user for edge-relation queries. Uses SEED_USER_ID
    ...    from the environment when set; otherwise creates one and marks it for cleanup.
    ${existing}=    Get Variable Value    ${SEED_USER_ID}    ${EMPTY}
    IF    $existing    RETURN
    ${email}=    Unique Email    seed.edge.user
    ${name}=    Unique Display Name    Seed Edge User
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SEED_USER_ID}    ${id}
    Set Global Variable    ${SEED_USER_CREATED}    ${True}

Ensure Seed Group
    [Documentation]    A group containing the seed user, so edge-column queries return
    ...    at least one group. Env-provided SEED_GROUP_ID is assumed to already have
    ...    members; a created one gets the seed user added.
    ${existing}=    Get Variable Value    ${SEED_GROUP_ID}    ${EMPTY}
    IF    $existing    RETURN
    Ensure Seed User
    ${gname}=    Unique Display Name    Seed Group
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $gname}, 'description': {'en-US': 'Seed group for API REST tests'}, 'owner_id': $SEED_USER_ID} }}
    ${gid}    ${getag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${SEED_GROUP_ID}    ${gid}
    Set Global Variable    ${SEED_GROUP_CREATED}    ${True}
    ${resp}=    POST On Session    api    /v1/iam/groups/${gid}/manage-users
    ...    json=${{ {'add': [$SEED_USER_ID]} }}
    Response Should Be Update Success    ${resp}    count=1

Ensure Archived User
    [Documentation]    An archived user for the include_archived search contract. A dedicated
    ...    fixture is required because 06_archive.robot deliberately unarchives the user under
    ...    test before 07_search.robot runs, so no archived row would otherwise exist.
    ...    The display name is stored too, so a search can be narrowed to this record alone.
    ${existing}=    Get Variable Value    ${ARCHIVED_USER_ID}    ${EMPTY}
    IF    $existing    RETURN
    ${email}=    Unique Email    seed.archived.user
    ${name}=    Unique Display Name    Seed Archived User
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    POST On Session    api    ${USER_API}/${id}/archived
    ...    json=${{ {'etag': $etag, 'is_archived': True} }}
    Response Should Be Update Success    ${resp}    count=1    previous_etag=${etag}
    Set Global Variable    ${ARCHIVED_USER_ID}    ${id}
    Set Global Variable    ${ARCHIVED_USER_NAME}    ${name}
    Set Global Variable    ${ARCHIVED_USER_CREATED}    ${True}

Create Probe User Session
    [Documentation]    Creates a user, issues a temporary password for them, signs in as
    ...    them and opens a session under the given alias. This is exactly what a human
    ...    administrator does during account recovery — the temp password comes back in
    ...    the response because the admin session holds manage_credentials on iam_user.
    ...    Returns the new user's id, email and session alias.
    [Arguments]    ${alias}
    ${email}=    Unique Email    perm.probe
    ${name}=    Unique Display Name    Perm Probe User
    ${resp}=    POST On Session    api    ${USER_API}
    ...    json=${{ {'display_name': $name, 'email': $email, 'status': 'active'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    users    ${id}

    # A user is created in "draft" and only an active account may hold a password,
    # so activation is part of provisioning exactly as it is for a real admin.
    ${resp}=    GET On Session    api    ${USER_API}/${id}
    ${item}=    Set Variable    ${resp.json()}[item]
    IF    '${item}[status]' != 'active'
        ${resp}=    PATCH On Session    api    ${USER_API}/${id}
        ...    json=${{ {'etag': $item['etag'], 'display_name': $name, 'email': $email, 'status': 'active'} }}
        ...    expected_status=any
        Should Be Equal As Integers    ${resp.status_code}    200
        ...    msg=Could not activate the probe user (status ${resp.status_code})
    END

    ${password}=    Issue Temp Password    ${email}
    Open Session As    ${alias}    ${email}    ${password}
    RETURN    ${id}    ${email}

Issue Temp Password
    [Documentation]    Issues a temporary password and returns it. Fails loudly when the
    ...    response carries no password: that means the caller lacks manage_credentials,
    ...    and every downstream assertion would otherwise fail for the wrong reason.
    ...
    ...    Pass expect_success=${False} when the call is meant to be refused (an inactive
    ...    or suspended account), in which case nothing is returned.
    [Arguments]    ${username}    ${expect_success}=${True}
    ${resp}=    POST On Session    api    ${TEMP_PASSWORD_API}
    ...    json=${{ {'subject_type': 'nikkiuser', 'username': $username} }}    expected_status=any
    IF    not $expect_success
        Should Be True    ${resp.status_code} >= 400
        ...    msg=A temp password was issued for an account that should not receive one
        RETURN    ${EMPTY}
    END
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    password
    ...    msg=No password returned; the admin session lacks manage_credentials on iam_user
    RETURN    ${resp.json()}[password]

Open Session As
    [Documentation]    Signs in as the given principal and opens a client-cert session
    ...    under the alias. Separate from Create Authorized API Session, which caches a
    ...    single admin token for the whole run.
    [Arguments]    ${alias}    ${username}    ${password}
    Create Anonymous API Session    alias=${alias}_signin
    ${resp}=    POST On Session    ${alias}_signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $username} }}
    ${attempt_id}=    Set Variable    ${resp.json()}[attempt_id]
    ${resp}=    POST On Session    ${alias}_signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $attempt_id, 'passwords': {'password': $password}} }}
    Should Be True    ${resp.json()}[done]    msg=Sign-in did not complete for ${username}
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    Authorization=Bearer ${resp.json()}[data][access_token]
    Create Client Cert Session    ${alias}    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}

Read Roles As
    [Documentation]    Lists roles as the given session. Used to check that the probe's
    ...    answer matches what the API actually allows.
    [Arguments]    ${alias}
    ${resp}=    GET On Session    ${alias}    ${ROLE_API}    expected_status=any
    RETURN    ${resp}

Response Should Be Permission Refusal
    [Documentation]    Asserts the response is an authorization refusal.
    ...
    ...    An insufficient-permission error answers 403 with `err_insufficient_permissions`
    ...    in the body; 401 is also accepted, since an unauthenticated caller is refused
    ...    before the permission is ever evaluated. Both status and body are asserted: the
    ...    status is what clients branch on, and the body is what makes the refusal
    ...    actionable by naming the missing entitlement.
    [Arguments]    ${resp}    ${what}
    Should Be True    ${resp.status_code} in (401, 403)
    ...    msg=${what} returned ${resp.status_code}; expected 403 (or 401 when unauthenticated)
    IF    ${resp.status_code} == 403
        Should Contain    ${resp.text}    err_insufficient_permissions
        ...    msg=${what} returned 403 for a reason other than permissions: ${resp.text}
    END

Probe Permission
    [Documentation]    Asks the permission probe as the session behind ${alias} and returns
    ...    the parsed body. Any non-200 is a failure here; the refusal cases assert the
    ...    status themselves with Probe Permission Raw.
    [Arguments]    ${alias}    ${expression}
    ${resp}=    POST On Session    ${alias}    ${TEST_PERM_API}
    ...    json=${{ {'expression': $expression} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}

Probe Permission Raw
    [Documentation]    Same call without status assertions, for the contract tests that
    ...    care about the status code itself.
    [Arguments]    ${alias}    ${expression}
    ${resp}=    POST On Session    ${alias}    ${TEST_PERM_API}
    ...    json=${{ {'expression': $expression} }}    expected_status=any
    RETURN    ${resp}

Permission Should Be Granted
    [Documentation]    Asserts the probe grants the expression, and that every match names
    ...    a real grant path. A match with no source id or no expression would mean the
    ...    provenance columns are not being populated.
    [Arguments]    ${alias}    ${expression}    ${expected_kind}=${EMPTY}    ${expected_source_id}=${EMPTY}
    ${body}=    Probe Permission    ${alias}    ${expression}
    Should Be True    ${body}[is_granted]    msg=Expected ${expression} to be granted
    Should Not Be Empty    ${body}[matches]    msg=Granted but no grant path was reported
    FOR    ${match}    IN    @{body}[matches]
        Should Not Be Empty    ${match}[source_id]
        Should Not Be Empty    ${match}[ent_expression]
    END
    IF    $expected_kind
        ${kinds}=    Evaluate    [m['source_kind'] for m in $body['matches']]
        Should Contain    ${kinds}    ${expected_kind}
    END
    IF    $expected_source_id
        ${ids}=    Evaluate    [m['source_id'] for m in $body['matches']]
        Should Contain    ${ids}    ${expected_source_id}
    END
    RETURN    ${body}

Permission Should Be Denied
    [Documentation]    Asserts the probe refuses, and that the refusal carries nothing
    ...    beyond "no" — a denial that leaked near-miss information would be a defect.
    [Arguments]    ${alias}    ${expression}
    ${body}=    Probe Permission    ${alias}    ${expression}
    Should Not Be True    ${body}[is_granted]    msg=Expected ${expression} to be denied
    Should Be Empty    ${body}[matches]    msg=A denial must not report any grant path

Resolve Resource And Action
    [Documentation]    Looks up a seeded resource by code and one of its actions by code,
    ...    returning both ids. Resolved at runtime rather than hard-coded so the suite
    ...    keeps working when the seed ULIDs change, and so it fails with a clear message
    ...    if a resource the tests assume is simply absent.
    [Arguments]    ${resource_code}    ${action_code}
    ${resp}=    GET On Session    api    /v1/iam/resources
    ...    params=${{ {'graph': '{"if":["code", "=", "' + $resource_code + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}    msg=No seeded resource with code "${resource_code}"
    ${resource_id}=    Set Variable    ${items}[0][id]

    ${resp}=    GET On Session    api    /v1/iam/resources/${resource_id}/actions
    Response Status Should Be    ${resp}    200
    ${actions}=    Set Variable    ${resp.json()}[items]
    ${matching}=    Evaluate    [a for a in $actions if a.get('code') == "${action_code}"]
    Should Not Be Empty    ${matching}
    ...    msg=Resource "${resource_code}" has no action with code "${action_code}"
    RETURN    ${resource_id}    ${matching}[0][id]

Create Probe Role
    [Documentation]    Creates a role carrying one entitlement, built from the action and
    ...    resource codes so the stored expression is the one the guard looks for.
    ...    Returns the role id.
    [Arguments]    ${label}    ${action_id}    ${resource_id}    ${scope}=domain    ${scope_field}=${EMPTY}    ${scope_id}=${EMPTY}
    ${rname}=    Unique Display Name    ${label}
    ${resp}=    POST On Session    api    ${ROLE_API}
    ...    json=${{ {'name': $rname, 'description': 'Robot permission suite role', 'is_requestable': False} }}
    ${role_id}    ${etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    roles    ${role_id}

    ${ent}=    Create Dictionary    name=${rname} ent    role_id=${role_id}
    ...    action_id=${action_id}    resource_id=${resource_id}    scope=${scope}
    IF    $scope_field
        Set To Dictionary    ${ent}    ${scope_field}=${scope_id}
    END
    ${resp}=    POST On Session    api    ${ENTITLEMENT_API}    json=${ent}
    ${ent_id}    ${ent_etag}=    Response Should Be Create Success    ${resp}
    Track Permission Fixture    entitlements    ${ent_id}
    RETURN    ${role_id}

Assign Role To User
    [Documentation]    Assigns a role directly to a user through the supported API.
    [Arguments]    ${user_id}    ${role_id}
    ${resp}=    POST On Session    api    /v1/iam/users/${user_id}/roles
    ...    json=${{ {'add': [$role_id]} }}
    Response Status Should Be    ${resp}    200

Unassign Role From User
    [Arguments]    ${user_id}    ${role_id}
    ${resp}=    POST On Session    api    /v1/iam/users/${user_id}/roles
    ...    json=${{ {'remove': [$role_id]} }}
    Response Status Should Be    ${resp}    200

Track Permission Fixture
    [Documentation]    Records a created record so the suite teardown can remove it.
    ...    Kept per-kind because deletion order matters (entitlements before roles).
    [Arguments]    ${kind}    ${id}
    ${fixtures}=    Get Variable Value    ${PERM_FIXTURES}    ${None}
    IF    $fixtures is None
        ${fixtures}=    Create Dictionary
        Set Global Variable    ${PERM_FIXTURES}    ${fixtures}
    END
    ${bucket}=    Get From Dictionary    ${fixtures}    ${kind}    default=${None}
    IF    $bucket is None
        ${bucket}=    Create List
        Set To Dictionary    ${fixtures}    ${kind}=${bucket}
    END
    Append To List    ${bucket}    ${id}

Delete Permission Fixtures
    [Documentation]    Best-effort teardown, in dependency order: entitlements, then roles,
    ...    then groups, then users. Every delete tolerates a missing record so a partially
    ...    failed run still cleans up what it did create.
    ${fixtures}=    Get Variable Value    ${PERM_FIXTURES}    ${None}
    IF    $fixtures is None    RETURN
    @{order}=    Create List    entitlements    roles    groups    users
    &{paths}=    Create Dictionary    entitlements=${ENTITLEMENT_API}    roles=${ROLE_API}
    ...    groups=${GROUP_API}    users=${USER_API}
    FOR    ${kind}    IN    @{order}
        ${bucket}=    Get From Dictionary    ${fixtures}    ${kind}    default=${None}
        IF    $bucket is None    CONTINUE
        ${path}=    Get From Dictionary    ${paths}    ${kind}
        FOR    ${id}    IN    @{bucket}
            DELETE On Session    api    ${path}/${id}    expected_status=any
        END
    END
    Set Global Variable    ${PERM_FIXTURES}    ${None}

Delete Seed Data
    [Documentation]    Best-effort removal of the on-the-fly seed group/user (in that
    ...    order — membership first). Env-provided seed entities are never deleted.
    ${archived_created}=    Get Variable Value    ${ARCHIVED_USER_CREATED}    ${False}
    IF    $archived_created
        DELETE On Session    api    ${USER_API}/${ARCHIVED_USER_ID}    expected_status=any
        Set Global Variable    ${ARCHIVED_USER_ID}    ${EMPTY}
        Set Global Variable    ${ARCHIVED_USER_CREATED}    ${False}
    END
    ${group_created}=    Get Variable Value    ${SEED_GROUP_CREATED}    ${False}
    IF    $group_created
        DELETE On Session    api    ${GROUP_API}/${SEED_GROUP_ID}    expected_status=any
        Set Global Variable    ${SEED_GROUP_ID}    ${EMPTY}
        Set Global Variable    ${SEED_GROUP_CREATED}    ${False}
    END
    ${user_created}=    Get Variable Value    ${SEED_USER_CREATED}    ${False}
    IF    $user_created
        DELETE On Session    api    ${USER_API}/${SEED_USER_ID}    expected_status=any
        Set Global Variable    ${SEED_USER_ID}    ${EMPTY}
        Set Global Variable    ${SEED_USER_CREATED}    ${False}
    END

Ensure Locale Ordering Groups
    [Documentation]    Creates three groups whose English and Vietnamese names sort in
    ...    deliberately OPPOSITE orders, plus one that has an English name only.
    ...
    ...    The inversion is the whole point: a search that merely happened to return rows
    ...    in some stable order would pass a single-locale assertion by luck. Ordering that
    ...    flips when the reader's language flips can only come from the server sorting on
    ...    the right jsonb key. All four share one suffix so the set sorts as a block
    ...    regardless of what else the database already holds.
    ${existing}=    Get Variable Value    ${LOCALE_GROUP_IDS}    ${None}
    IF    $existing is not None    RETURN
    Ensure Seed User
    ${suffix}=    Unique Suffix

    # en-US ascending: Alpha, Bravo, Charlie
    # vi-VN ascending: Xuan(Charlie), Yen(Bravo), Zulu(Alpha)  -> exactly reversed
    @{specs}=    Create List
    ...    ${{ {'key': 'alpha',   'en': 'Alpha ' + $suffix,   'vi': 'Zulu ' + $suffix} }}
    ...    ${{ {'key': 'bravo',   'en': 'Bravo ' + $suffix,   'vi': 'Yen ' + $suffix} }}
    ...    ${{ {'key': 'charlie', 'en': 'Charlie ' + $suffix, 'vi': 'Xuan ' + $suffix} }}

    &{ids}=    Create Dictionary
    FOR    ${spec}    IN    @{specs}
        ${resp}=    POST On Session    api    ${GROUP_API}
        ...    json=${{ {'name': {'en-US': $spec['en'], 'vi-VN': $spec['vi']}, 'owner_id': $SEED_USER_ID} }}
        ${id}    ${etag}=    Response Should Be Create Success    ${resp}
        Set To Dictionary    ${ids}    ${spec}[key]    ${id}
    END

    # English only: it has nothing to match or sort on for a Vietnamese reader, which is
    # what pins the strict no-fallback rule.
    ${en_only}=    Set Variable    Delta ${suffix}
    ${resp}=    POST On Session    api    ${GROUP_API}
    ...    json=${{ {'name': {'en-US': $en_only}, 'owner_id': $SEED_USER_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set To Dictionary    ${ids}    en_only    ${id}

    Set Global Variable    ${LOCALE_GROUP_IDS}    ${ids}
    Set Global Variable    ${LOCALE_GROUP_SUFFIX}    ${suffix}

Delete Locale Group Fixtures
    [Documentation]    Best-effort removal of the groups Ensure Locale Ordering Groups made.
    ${ids}=    Get Variable Value    ${LOCALE_GROUP_IDS}    ${None}
    IF    $ids is None    RETURN
    FOR    ${key}    ${id}    IN    &{ids}
        DELETE On Session    api    ${GROUP_API}/${id}    expected_status=any
    END
    Set Global Variable    ${LOCALE_GROUP_IDS}    ${None}

Locale Group Names In Result
    [Documentation]    The returned group names in one language, keeping result order and
    ...    dropping anything this suite did not create -- the database may hold groups from
    ...    other suites, and they must not decide whether an ordering assertion passes.
    [Arguments]    ${resp}    ${language}
    ${names}=    Evaluate
    ...    [i['name'].get($language) for i in $resp.json()['items'] if isinstance(i.get('name'), dict) and $LOCALE_GROUP_SUFFIX in str(i['name'])]
    RETURN    ${names}
