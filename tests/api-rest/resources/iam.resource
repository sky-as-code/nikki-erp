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
${IAM_SCHEMA_DIR}       ${CURDIR}/schemas/iam


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
